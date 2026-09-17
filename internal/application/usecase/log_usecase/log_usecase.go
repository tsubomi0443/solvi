package log_usecase

import (
	"crypto/rand"
	"fmt"
	"io"
	"solvi/internal/domain/entity"
	"solvi/internal/domain/interface/repository"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yeka/zip"
)

const ticketTTL = 10 * time.Minute

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

func (uc *LogUsecase) ListAvailableDates() []string {
	dates, err := uc.logRepository.ListDates()
	if err != nil {
		return []string{}
	}
	return dates
}

func (uc *LogUsecase) IssueAll(userUUID string) (*IssueResult, error) {
	names, err := uc.logRepository.ListNames()
	if err != nil {
		return nil, fmt.Errorf("ログファイルの取得に失敗しました: %w", err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("ダウンロード可能なログがありません")
	}

	return uc.issueTicket(userUUID, entity.DownloadKindAll, time.Time{}, "solvi-logs-all.zip")
}

func (uc *LogUsecase) IssueDate(userUUID string, date string) (*IssueResult, error) {
	parsed, err := parseDate(date)
	if err != nil {
		return nil, err
	}
	names, err := uc.logRepository.ListByDate(parsed)
	if err != nil {
		return nil, fmt.Errorf("ログファイルの取得に失敗しました: %w", err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("対象日付のログファイルが見つかりません")
	}

	filename := fmt.Sprintf("solvi-logs-%s.zip", parsed.Format("20060102"))
	return uc.issueTicket(userUUID, entity.DownloadKindDate, parsed, filename)
}

func (uc *LogUsecase) LookupTicket(key, userUUID string) (entity.DownloadTicket, error) {
	return uc.lookupTicket(key, userUUID)
}

func (uc *LogUsecase) Stream(writer io.Writer, key, userUUID string) error {
	ticket, err := uc.lookupTicket(key, userUUID)
	if err != nil {
		return err
	}

	names, err := uc.resolveNames(ticket)
	if err != nil {
		return err
	}

	if err := uc.streamZip(writer, names, ticket.Password); err != nil {
		return fmt.Errorf("ログファイルの暗号圧縮に失敗しました: %w", err)
	}

	uc.deleteTicket(key)
	return nil
}

func (uc *LogUsecase) issueTicket(userUUID string, kind entity.DownloadKind, date time.Time, filename string) (*IssueResult, error) {
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

func (uc *LogUsecase) resolveNames(ticket entity.DownloadTicket) ([]string, error) {
	switch ticket.Kind {
	case entity.DownloadKindAll:
		names, err := uc.logRepository.ListNames()
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("ダウンロード可能なログがありません")
		}
		return names, nil
	case entity.DownloadKindDate:
		names, err := uc.logRepository.ListByDate(ticket.Date)
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

func (uc *LogUsecase) streamZip(writer io.Writer, names []string, password string) error {
	zipWriter := zip.NewWriter(writer)
	for _, name := range names {
		rc, err := uc.logRepository.Open(name)
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
