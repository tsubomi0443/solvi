package log_usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"solvi/internal/application/usecase"
	"solvi/internal/domain/entity"
	"solvi/internal/domain/interface/repository"
	"sync"
	"time"

	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
	"github.com/yeka/zip"
)

const (
	ticketTTL = 10 * time.Minute
	opLog     = "log_usecase"
)

type IssueResult struct {
	DownloadKey string `json:"downloadKey"`
	Password    string `json:"password"`
	Filename    string `json:"filename"`
}

type LogUsecase struct {
	logRepository repository.LogRepository
	tickets       map[string]entity.DownloadTicket
	mu            sync.Mutex
}

func NewLogUsecase(logRepo repository.LogRepository) *LogUsecase {
	return &LogUsecase{
		logRepository: logRepo,
		tickets:       make(map[string]entity.DownloadTicket),
	}
}

func (uc *LogUsecase) ListAvailableDates(ctx context.Context) []string {
	const op = opLog + ".ListAvailableDates"
	dates, err := uc.logRepository.ListDates(ctx)
	if err != nil {
		logutils.Warn(ctx, logutils.LayerUsecase, op, "日付一覧取得失敗、空配列を返却", slog.String("err", err.Error()))
		return []string{}
	}
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理完了", slog.Int("count", len(dates)))
	return dates
}

func (uc *LogUsecase) IssueAll(ctx context.Context, userUUID string) (*IssueResult, error) {
	const op = opLog + ".IssueAll"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("user_uuid", userUUID))
	names, err := uc.logRepository.ListNames(ctx)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ログファイル一覧取得失敗", err, slog.String("user_uuid", userUUID))
		return nil, fmt.Errorf("ログファイルの取得に失敗しました: %w", err)
	}
	if len(names) == 0 {
		usecase.LogBusinessWarn(ctx, op, "ダウンロード可能なログなし", fmt.Errorf("ダウンロード可能なログがありません"), slog.String("user_uuid", userUUID))
		return nil, fmt.Errorf("ダウンロード可能なログがありません")
	}

	result, err := uc.issueTicket(ctx, userUUID, entity.DownloadKindAll, time.Time{}, "solvi-logs-all.zip")
	if err != nil {
		return nil, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログダウンロード発行成功", slog.String("user_uuid", userUUID), slog.String("download_key", result.DownloadKey))
	return result, nil
}

func (uc *LogUsecase) IssueDate(ctx context.Context, userUUID string, date string) (*IssueResult, error) {
	const op = opLog + ".IssueDate"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("user_uuid", userUUID), slog.String("date", date))
	parsed, err := parseDate(date)
	if err != nil {
		usecase.LogBusinessWarn(ctx, op, "日付形式不正", err, slog.String("date", date))
		return nil, err
	}
	names, err := uc.logRepository.ListByDate(ctx, parsed)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ログファイル取得失敗", err, slog.String("user_uuid", userUUID), slog.String("date", date))
		return nil, fmt.Errorf("ログファイルの取得に失敗しました: %w", err)
	}
	if len(names) == 0 {
		usecase.LogBusinessWarn(ctx, op, "対象日付のログなし", fmt.Errorf("対象日付のログファイルが見つかりません"), slog.String("date", date))
		return nil, fmt.Errorf("対象日付のログファイルが見つかりません")
	}

	filename := fmt.Sprintf("solvi-logs-%s.zip", parsed.Format("20060102"))
	result, err := uc.issueTicket(ctx, userUUID, entity.DownloadKindDate, parsed, filename)
	if err != nil {
		return nil, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログダウンロード発行成功", slog.String("user_uuid", userUUID), slog.String("date", date), slog.String("download_key", result.DownloadKey))
	return result, nil
}

func (uc *LogUsecase) LookupTicket(ctx context.Context, key, userUUID string) (entity.DownloadTicket, error) {
	const op = opLog + ".LookupTicket"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "チケット照会", slog.String("key", key), slog.String("user_uuid", userUUID))
	ticket, err := uc.lookupTicket(key, userUUID)
	if err != nil {
		usecase.LogBusinessWarn(ctx, op, "チケット照会失敗", err, slog.String("key", key), slog.String("user_uuid", userUUID))
	}
	return ticket, err
}

func (uc *LogUsecase) Stream(ctx context.Context, writer io.Writer, key, userUUID string) error {
	const op = opLog + ".Stream"
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログストリーム開始", slog.String("key", key), slog.String("user_uuid", userUUID))
	ticket, err := uc.lookupTicket(key, userUUID)
	if err != nil {
		usecase.LogBusinessWarn(ctx, op, "チケット照会失敗", err, slog.String("key", key), slog.String("user_uuid", userUUID))
		return err
	}

	names, err := uc.resolveNames(ctx, ticket)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "ログファイル解決失敗", err, slog.String("key", key))
		return err
	}

	if err := uc.streamZip(ctx, writer, names, ticket.Password); err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "ログストリーム失敗", slog.String("key", key), slog.String("user_uuid", userUUID), slog.String("err", err.Error()))
		return fmt.Errorf("ログファイルの暗号圧縮に失敗しました: %w", err)
	}

	uc.deleteTicket(key)
	logutils.Info(ctx, logutils.LayerUsecase, op, "ログストリーム完了", slog.String("key", key), slog.String("user_uuid", userUUID), slog.String("filename", ticket.Filename))
	return nil
}

func (uc *LogUsecase) issueTicket(ctx context.Context, userUUID string, kind entity.DownloadKind, date time.Time, filename string) (*IssueResult, error) {
	uc.cleanupExpiredTickets()

	key := uuid.NewString()
	password := genpw(32)
	ticket := entity.DownloadTicket{
		Key:       key,
		Password:  password,
		UserUUID:  userUUID,
		Kind:      kind,
		Date:      date,
		ExpiresAt: time.Now().Add(ticketTTL),
		Filename:  filename,
	}

	uc.mu.Lock()
	uc.tickets[key] = ticket
	uc.mu.Unlock()

	logutils.Debug(ctx, logutils.LayerUsecase, opLog+".issueTicket", "チケット発行", slog.String("key", key), slog.String("user_uuid", userUUID))
	return &IssueResult{
		DownloadKey: key,
		Password:    password,
		Filename:    filename,
	}, nil
}

func (uc *LogUsecase) lookupTicket(key, userUUID string) (entity.DownloadTicket, error) {
	uc.cleanupExpiredTickets()

	uc.mu.Lock()
	ticket, ok := uc.tickets[key]
	uc.mu.Unlock()
	if !ok {
		return entity.DownloadTicket{}, fmt.Errorf("ダウンロードキーが無効です")
	}
	if ticket.UserUUID != userUUID {
		return entity.DownloadTicket{}, fmt.Errorf("ダウンロードキーにアクセスできません")
	}
	if time.Now().After(ticket.ExpiresAt) {
		return entity.DownloadTicket{}, fmt.Errorf("ダウンロードキーの有効期限が切れています")
	}

	return ticket, nil
}

func (uc *LogUsecase) deleteTicket(key string) {
	uc.mu.Lock()
	delete(uc.tickets, key)
	uc.mu.Unlock()
}

func (uc *LogUsecase) cleanupExpiredTickets() {
	now := time.Now()
	uc.mu.Lock()
	for key, ticket := range uc.tickets {
		if now.After(ticket.ExpiresAt) {
			delete(uc.tickets, key)
		}
	}
	uc.mu.Unlock()
}

func (uc *LogUsecase) resolveNames(ctx context.Context, ticket entity.DownloadTicket) ([]string, error) {
	switch ticket.Kind {
	case entity.DownloadKindAll:
		names, err := uc.logRepository.ListNames(ctx)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("ダウンロード可能なログがありません")
		}
		return names, nil
	case entity.DownloadKindDate:
		names, err := uc.logRepository.ListByDate(ctx, ticket.Date)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("対象日付のログファイルが見つかりません")
		}
		return names, nil
	default:
		return nil, fmt.Errorf("不明なダウンロード種別です")
	}
}

func (uc *LogUsecase) streamZip(ctx context.Context, writer io.Writer, names []string, password string) error {
	zipWriter := zip.NewWriter(writer)
	for _, name := range names {
		rc, err := uc.logRepository.Open(ctx, name)
		if err != nil {
			zipWriter.Close()
			return err
		}

		zipFile, err := zipWriter.Encrypt(name, password, zip.AES256Encryption)
		if err != nil {
			rc.Close()
			zipWriter.Close()
			return fmt.Errorf("暗号化Zipファイルの生成に失敗しました: %w", err)
		}

		if _, err := io.Copy(zipFile, rc); err != nil {
			rc.Close()
			zipWriter.Close()
			return fmt.Errorf("ログファイルの圧縮に失敗しました: %w", err)
		}
		if err := rc.Close(); err != nil {
			zipWriter.Close()
			return fmt.Errorf("ログファイルのクローズに失敗しました: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("Zipファイルの終了処理に失敗しました: %w", err)
	}

	return nil
}

func parseDate(date string) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("日付の形式が不正です")
	}
	return parsed, nil
}

const (
	PW_MAX_LEN = 26
	PW_MIN_LEN = 12
)

func genpw(length uint) string {
	l := max(min(length, PW_MAX_LEN), PW_MIN_LEN)
	return rand.Text()[:l]
}
