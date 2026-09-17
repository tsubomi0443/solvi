package question_usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"solvi/internal/application/converter"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/application/usecase"
	"solvi/internal/domain/entity"
	ext "solvi/internal/domain/interface/external"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/domain/valueobject"
	"solvi/internal/shared/config"
	logutils "solvi/internal/shared/logUtils"

	"gorm.io/gorm"
)

const opQuestion = "question_usecase"

type QuestionUsecase struct {
	questionRepo repo.QuestionRepository
	userRepo     repo.UserRepository
	bedrock      ext.BedrockClient
	onAIAnswer   func(questionUUID string)
}

func NewQuestionUsecase(q repo.QuestionRepository, u repo.UserRepository, b ext.BedrockClient, onAIAnswer func(string)) *QuestionUsecase {
	return &QuestionUsecase{questionRepo: q, userRepo: u, bedrock: b, onAIAnswer: onAIAnswer}
}

func canViewAllQuestions(isSupporter, isAdmin bool) bool {
	return isSupporter || isAdmin
}

func (uc *QuestionUsecase) List(ctx context.Context, actorID uint, isSupporter, isAdmin bool) ([]outputmodel.QuestionListItemOutput, error) {
	const op = opQuestion + ".List"
	if canViewAllQuestions(isSupporter, isAdmin) {
		logutils.Debug(ctx, logutils.LayerUsecase, op, "全件一覧取得", slog.Uint64("actor_id", uint64(actorID)))
		qs, err := uc.questionRepo.ListAll(ctx)
		if err != nil {
			usecase.LogRepoPropagation(ctx, op, "質問一覧取得失敗", err, slog.Uint64("actor_id", uint64(actorID)))
			return nil, err
		}
		return mapList(qs), nil
	}
	logutils.Debug(ctx, logutils.LayerUsecase, op, "自分の質問一覧取得", slog.Uint64("actor_id", uint64(actorID)))
	qs, err := uc.questionRepo.ListByQuestionUserID(ctx, actorID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問一覧取得失敗", err, slog.Uint64("actor_id", uint64(actorID)))
		return nil, err
	}
	return mapList(qs), nil
}

func mapList(qs []entity.Question) []outputmodel.QuestionListItemOutput {
	out := make([]outputmodel.QuestionListItemOutput, 0, len(qs))
	for _, q := range qs {
		out = append(out, converter.QuestionEntityToListItem(&q))
	}
	return out
}

func (uc *QuestionUsecase) Get(ctx context.Context, actorID uint, isSupporter, isAdmin bool, uuid string) (*outputmodel.QuestionDetailOutput, error) {
	const op = opQuestion + ".Get"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			usecase.LogBusinessWarn(ctx, op, "質問が見つかりません", err, slog.String("question_uuid", uuid))
		} else {
			usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		}
		return nil, err
	}
	if !canViewAllQuestions(isSupporter, isAdmin) && q.QuestionUserID != actorID {
		usecase.LogBusinessWarn(ctx, op, "閲覧権限なし", fmt.Errorf("閲覧権限がありません"), slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
		return nil, fmt.Errorf("閲覧権限がありません")
	}
	return new(converter.QuestionEntityToDetail(q, canViewAllQuestions(isSupporter, isAdmin))), nil
}

func (uc *QuestionUsecase) Create(ctx context.Context, actorID uint, title, content string, tags []string, answerDue *time.Time, requireHuman bool) (*outputmodel.QuestionDetailOutput, error) {
	const op = opQuestion + ".Create"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.Uint64("actor_id", uint64(actorID)), slog.Int("content_len", len(content)), slog.Int("tag_count", len(tags)))
	q := &entity.Question{
		Title:                 title,
		IsRequireHumanSupport: requireHuman,
		SupportStatus:         valueobject.SupportStatusPending,
		QuestionUserID:        actorID,
		AnswerDue:             answerDue,
		Contents:              []entity.QuestionContent{{Content: content, QuestionUserID: actorID}},
	}
	for _, t := range tags {
		if s := strings.TrimSpace(t); s != "" {
			q.Tags = append(q.Tags, entity.QuestionTag{Name: s})
		}
	}
	if err := uc.questionRepo.Create(ctx, q); err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問作成失敗", err, slog.Uint64("actor_id", uint64(actorID)))
		return nil, err
	}
	created, err := uc.questionRepo.GetByUUID(ctx, q.UUID.String())
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "作成後の質問取得失敗", err, slog.String("question_uuid", q.UUID.String()))
		return nil, err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "質問作成成功", slog.String("question_uuid", created.UUID.String()), slog.Uint64("actor_id", uint64(actorID)))
	return new(converter.QuestionEntityToDetail(created, false)), nil
}

func (uc *QuestionUsecase) runAI(questionUUID, title, content string) {
	result, err := uc.bedrock.AnswerQuestion(title, content)
	if err != nil {
		slog.Error("AI回答失敗", "err", err, "uuid", questionUUID)
		q, e := uc.questionRepo.GetByUUID(context.Background(), questionUUID)
		if e == nil {
			q.IsRequireHumanSupport = true
			_ = uc.questionRepo.Update(context.Background(), q)
		}
		return
	}
	sys, err := uc.userRepo.GetByEmail(context.Background(), config.GetSystemUserEmail())
	if err != nil {
		slog.Error("system user not found", "err", err)
		return
	}
	q, err := uc.questionRepo.GetByUUID(context.Background(), questionUUID)
	if err != nil {
		return
	}
	_ = uc.questionRepo.AddAnswer(context.Background(), &entity.QuestionAnswer{Content: result.Content, AnswerUserID: sys.ID, QuestionID: q.ID})
	for _, ref := range result.References {
		_ = uc.questionRepo.AddRefer(context.Background(), &entity.QuestionRefer{Name: ref.Name, URL: ref.URL, QuestionID: q.ID, UserID: sys.ID})
	}
	if uc.onAIAnswer != nil {
		uc.onAIAnswer(questionUUID)
	}
}

func (uc *QuestionUsecase) AppendContent(ctx context.Context, actorID uint, uuid, content string) error {
	const op = opQuestion + ".AppendContent"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Int("content_len", len(content)))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	if q.QuestionUserID != actorID {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
		return fmt.Errorf("権限がありません")
	}
	if err := uc.questionRepo.AddContent(ctx, &entity.QuestionContent{Content: content, QuestionUserID: actorID, QuestionID: q.ID}); err != nil {
		usecase.LogRepoPropagation(ctx, op, "追記失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "追記成功", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	return nil
}

func (uc *QuestionUsecase) AddAnswer(ctx context.Context, actorID uint, uuid, content string, refers []outputmodel.ReferOutput) error {
	const op = opQuestion + ".AddAnswer"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	if err := uc.questionRepo.AddAnswer(ctx, &entity.QuestionAnswer{Content: content, AnswerUserID: actorID, QuestionID: q.ID}); err != nil {
		usecase.LogRepoPropagation(ctx, op, "回答追加失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	for _, r := range refers {
		if err := uc.questionRepo.AddRefer(ctx, &entity.QuestionRefer{Name: r.Name, URL: r.URL, QuestionID: q.ID, UserID: actorID}); err != nil {
			logutils.Warn(ctx, logutils.LayerUsecase, op, "引用追加失敗", slog.String("question_uuid", uuid), slog.String("err", err.Error()))
		}
	}
	if q.SupportStatus == valueobject.SupportStatusPending {
		q.SupportStatus = valueobject.SupportStatusSupporting
		if err := uc.questionRepo.Update(ctx, q); err != nil {
			logutils.Warn(ctx, logutils.LayerUsecase, op, "ステータス更新失敗", slog.String("question_uuid", uuid), slog.String("err", err.Error()))
		}
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "回答追加成功", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	return nil
}

func (uc *QuestionUsecase) AddRefer(ctx context.Context, actorID uint, uuid, name, url string) error {
	return uc.AddRefers(ctx, actorID, uuid, []outputmodel.ReferOutput{{Name: name, URL: url}})
}

func (uc *QuestionUsecase) AddRefers(ctx context.Context, actorID uint, uuid string, refers []outputmodel.ReferOutput) error {
	const op = opQuestion + ".AddRefers"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Int("refer_count", len(refers)))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	valid := make([]outputmodel.ReferOutput, 0, len(refers))
	for _, r := range refers {
		name := strings.TrimSpace(r.Name)
		url := strings.TrimSpace(r.URL)
		if name == "" && url == "" {
			continue
		}
		if name == "" || url == "" {
			usecase.LogBusinessWarn(ctx, op, "引用入力不正", fmt.Errorf("タイトルとURLは両方入力してください"), slog.String("question_uuid", uuid))
			return fmt.Errorf("タイトルとURLは両方入力してください")
		}
		valid = append(valid, outputmodel.ReferOutput{Name: name, URL: url})
	}
	if len(valid) == 0 {
		usecase.LogBusinessWarn(ctx, op, "引用情報なし", fmt.Errorf("引用情報がありません"), slog.String("question_uuid", uuid))
		return fmt.Errorf("引用情報がありません")
	}
	for _, r := range valid {
		if err := uc.questionRepo.AddRefer(ctx, &entity.QuestionRefer{
			Name: r.Name, URL: r.URL, QuestionID: q.ID, UserID: actorID,
		}); err != nil {
			usecase.LogRepoPropagation(ctx, op, "引用追加失敗", err, slog.String("question_uuid", uuid))
			return err
		}
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "引用追加成功", slog.String("question_uuid", uuid), slog.Int("count", len(valid)))
	return nil
}

func (uc *QuestionUsecase) AddMemo(ctx context.Context, actorID uint, uuid, content string) error {
	const op = opQuestion + ".AddMemo"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	if err := uc.questionRepo.AddMemo(ctx, &entity.QuestionMemo{Content: content, QuestionID: q.ID, MemoUserID: actorID}); err != nil {
		usecase.LogRepoPropagation(ctx, op, "メモ追加失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "メモ追加成功", slog.String("question_uuid", uuid), slog.Uint64("actor_id", uint64(actorID)))
	return nil
}

func (uc *QuestionUsecase) DeleteAnswer(ctx context.Context, actorID uint, isSupporter, isAdmin bool, questionUUID, answerUUID string) error {
	const op = opQuestion + ".DeleteAnswer"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", questionUUID), slog.String("answer_uuid", answerUUID))
	if !isSupporter && !isAdmin {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", questionUUID))
		return fmt.Errorf("権限がありません")
	}
	q, err := uc.questionRepo.GetByUUID(ctx, questionUUID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", questionUUID))
		return err
	}
	var target *entity.QuestionAnswer
	for i := range q.Answers {
		if q.Answers[i].UUID.String() == answerUUID {
			target = &q.Answers[i]
			break
		}
	}
	if target == nil {
		usecase.LogBusinessWarn(ctx, op, "回答が見つかりません", fmt.Errorf("回答が見つかりません"), slog.String("answer_uuid", answerUUID))
		return fmt.Errorf("回答が見つかりません")
	}
	if !isAdmin && target.AnswerUserID != actorID {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("answer_uuid", answerUUID))
		return fmt.Errorf("権限がありません")
	}
	if err := uc.questionRepo.SoftDeleteAnswerByUUID(ctx, answerUUID); err != nil {
		usecase.LogRepoPropagation(ctx, op, "回答削除失敗", err, slog.String("answer_uuid", answerUUID))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "回答削除成功", slog.String("question_uuid", questionUUID), slog.String("answer_uuid", answerUUID))
	return nil
}

func (uc *QuestionUsecase) DeleteMemo(ctx context.Context, actorID uint, isSupporter, isAdmin bool, questionUUID, memoUUID string) error {
	const op = opQuestion + ".DeleteMemo"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", questionUUID), slog.String("memo_uuid", memoUUID))
	if !isSupporter && !isAdmin {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", questionUUID))
		return fmt.Errorf("権限がありません")
	}
	q, err := uc.questionRepo.GetByUUID(ctx, questionUUID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", questionUUID))
		return err
	}
	var target *entity.QuestionMemo
	for i := range q.Memos {
		if q.Memos[i].UUID.String() == memoUUID {
			target = &q.Memos[i]
			break
		}
	}
	if target == nil {
		usecase.LogBusinessWarn(ctx, op, "メモが見つかりません", fmt.Errorf("メモが見つかりません"), slog.String("memo_uuid", memoUUID))
		return fmt.Errorf("メモが見つかりません")
	}
	if !isAdmin && target.MemoUserID != actorID {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("memo_uuid", memoUUID))
		return fmt.Errorf("権限がありません")
	}
	if err := uc.questionRepo.SoftDeleteMemoByUUID(ctx, memoUUID); err != nil {
		usecase.LogRepoPropagation(ctx, op, "メモ削除失敗", err, slog.String("memo_uuid", memoUUID))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "メモ削除成功", slog.String("question_uuid", questionUUID), slog.String("memo_uuid", memoUUID))
	return nil
}

func (uc *QuestionUsecase) DeleteRefer(ctx context.Context, actorID uint, isSupporter, isAdmin bool, questionUUID, referUUID string) error {
	const op = opQuestion + ".DeleteRefer"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", questionUUID), slog.String("refer_uuid", referUUID))
	if !isSupporter && !isAdmin {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", questionUUID))
		return fmt.Errorf("権限がありません")
	}
	q, err := uc.questionRepo.GetByUUID(ctx, questionUUID)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", questionUUID))
		return err
	}
	var target *entity.QuestionRefer
	for i := range q.Refers {
		if q.Refers[i].UUID.String() == referUUID {
			target = &q.Refers[i]
			break
		}
	}
	if target == nil {
		usecase.LogBusinessWarn(ctx, op, "引用情報が見つかりません", fmt.Errorf("引用情報が見つかりません"), slog.String("refer_uuid", referUUID))
		return fmt.Errorf("引用情報が見つかりません")
	}
	if !isAdmin && target.UserID != actorID {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("refer_uuid", referUUID))
		return fmt.Errorf("権限がありません")
	}
	if err := uc.questionRepo.SoftDeleteReferByUUID(ctx, referUUID); err != nil {
		usecase.LogRepoPropagation(ctx, op, "引用削除失敗", err, slog.String("refer_uuid", referUUID))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "引用削除成功", slog.String("question_uuid", questionUUID), slog.String("refer_uuid", referUUID))
	return nil
}

func (uc *QuestionUsecase) Delete(ctx context.Context, isAdmin bool, uuid string) error {
	const op = opQuestion + ".Delete"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid))
	if !isAdmin {
		usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", uuid))
		return fmt.Errorf("権限がありません")
	}
	if _, err := uc.questionRepo.GetByUUID(ctx, uuid); err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	if err := uc.questionRepo.SoftDeleteByUUID(ctx, uuid); err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問削除失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "質問削除成功", slog.String("question_uuid", uuid))
	return nil
}

type QuestionSummaryInput struct {
	Content    string
	Answer     string
	ReferUUIDs []string
}

func (uc *QuestionUsecase) Update(ctx context.Context, actorID uint, isAdmin, isSupporter bool, uuid string, title *string, status *string, due *time.Time, tags *[]string, requireHuman *bool, complete bool, summary *QuestionSummaryInput) error {
	const op = opQuestion + ".Update"
	logutils.Debug(ctx, logutils.LayerUsecase, op, "処理開始", slog.String("question_uuid", uuid), slog.Bool("complete", complete))
	q, err := uc.questionRepo.GetByUUID(ctx, uuid)
	if err != nil {
		usecase.LogRepoPropagation(ctx, op, "質問取得失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	if complete {
		if isSupporter || (!q.IsRequireHumanSupport && q.QuestionUserID == actorID) {
			q.SupportStatus = valueobject.SupportStatusDone
			if err := uc.questionRepo.Update(ctx, q); err != nil {
				usecase.LogRepoPropagation(ctx, op, "完了更新失敗", err, slog.String("question_uuid", uuid))
				return err
			}
			if summary != nil {
				if err := uc.upsertSummary(ctx, q, *summary); err != nil {
					usecase.LogRepoPropagation(ctx, op, "サマリー保存失敗", err, slog.String("question_uuid", uuid))
					return err
				}
			} else {
				if err := uc.createSummary(ctx, q); err != nil {
					usecase.LogRepoPropagation(ctx, op, "サマリー作成失敗", err, slog.String("question_uuid", uuid))
					return err
				}
			}
			logutils.Info(ctx, logutils.LayerUsecase, op, "質問完了", slog.String("question_uuid", uuid))
			return nil
		}
		usecase.LogBusinessWarn(ctx, op, "完了不可", fmt.Errorf("完了できません"), slog.String("question_uuid", uuid))
		return fmt.Errorf("完了できません")
	}
	if !isSupporter && !isAdmin {
		if q.QuestionUserID != actorID {
			usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", uuid))
			return fmt.Errorf("権限がありません")
		}
		if title != nil || status != nil || due != nil || tags != nil {
			usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", uuid))
			return fmt.Errorf("権限がありません")
		}
		if requireHuman == nil {
			usecase.LogBusinessWarn(ctx, op, "権限なし", fmt.Errorf("権限がありません"), slog.String("question_uuid", uuid))
			return fmt.Errorf("権限がありません")
		}
		q.IsRequireHumanSupport = *requireHuman
		if err := uc.questionRepo.Update(ctx, q); err != nil {
			usecase.LogRepoPropagation(ctx, op, "更新失敗", err, slog.String("question_uuid", uuid))
			return err
		}
		logutils.Info(ctx, logutils.LayerUsecase, op, "質問更新成功", slog.String("question_uuid", uuid))
		return nil
	}
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			usecase.LogBusinessWarn(ctx, op, "タイトル必須", fmt.Errorf("タイトルは必須です"), slog.String("question_uuid", uuid))
			return fmt.Errorf("タイトルは必須です")
		}
		q.Title = trimmed
	}
	if status != nil {
		parsed, err := valueobject.ParseSupportStatus(supportStatusToInt(*status))
		if err != nil {
			usecase.LogBusinessWarn(ctx, op, "ステータス不正", err, slog.String("question_uuid", uuid))
			return err
		}
		if parsed == valueobject.SupportStatusDone {
			if summary == nil {
				usecase.LogBusinessWarn(ctx, op, "サマリー必須", fmt.Errorf("完了に変更する場合は要約が必要です"), slog.String("question_uuid", uuid))
				return fmt.Errorf("完了に変更する場合は要約が必要です")
			}
			if strings.TrimSpace(summary.Content) == "" || strings.TrimSpace(summary.Answer) == "" {
				usecase.LogBusinessWarn(ctx, op, "要約未入力", fmt.Errorf("質問の要約と対応の要約は必須です"), slog.String("question_uuid", uuid))
				return fmt.Errorf("質問の要約と対応の要約は必須です")
			}
			if len(q.Refers) > 0 && len(summary.ReferUUIDs) == 0 {
				usecase.LogBusinessWarn(ctx, op, "引用未選択", fmt.Errorf("引用を1件以上選択してください"), slog.String("question_uuid", uuid))
				return fmt.Errorf("引用を1件以上選択してください")
			}
			q.SupportStatus = parsed
			if err := uc.questionRepo.Update(ctx, q); err != nil {
				usecase.LogRepoPropagation(ctx, op, "更新失敗", err, slog.String("question_uuid", uuid))
				return err
			}
			if err := uc.upsertSummary(ctx, q, *summary); err != nil {
				usecase.LogRepoPropagation(ctx, op, "サマリー保存失敗", err, slog.String("question_uuid", uuid))
				return err
			}
			logutils.Info(ctx, logutils.LayerUsecase, op, "質問完了", slog.String("question_uuid", uuid))
			return nil
		}
		q.SupportStatus = parsed
	}
	if due != nil {
		q.AnswerDue = due
	}
	if requireHuman != nil {
		q.IsRequireHumanSupport = *requireHuman
	}
	if tags != nil {
		tagEntities := make([]entity.QuestionTag, 0, len(*tags))
		for _, t := range *tags {
			if s := strings.TrimSpace(t); s != "" {
				tagEntities = append(tagEntities, entity.QuestionTag{Name: s, QuestionID: q.ID})
			}
		}
		if err := uc.questionRepo.ReplaceTags(ctx, q.ID, tagEntities); err != nil {
			usecase.LogRepoPropagation(ctx, op, "タグ更新失敗", err, slog.String("question_uuid", uuid))
			return err
		}
	}
	if err := uc.questionRepo.Update(ctx, q); err != nil {
		usecase.LogRepoPropagation(ctx, op, "更新失敗", err, slog.String("question_uuid", uuid))
		return err
	}
	logutils.Info(ctx, logutils.LayerUsecase, op, "質問更新成功", slog.String("question_uuid", uuid))
	return nil
}

func supportStatusToInt(s string) int {
	switch s {
	case "pending":
		return 1
	case "supporting":
		return 2
	case "done":
		return 3
	default:
		return 0
	}
}

func (uc *QuestionUsecase) createSummary(ctx context.Context, q *entity.Question) error {
	full, err := uc.questionRepo.GetByUUID(ctx, q.UUID.String())
	if err != nil {
		return err
	}
	body := ""
	for _, c := range full.Contents {
		body += c.Content + "\n"
	}
	answer := ""
	for _, a := range full.Answers {
		answer = a.Content
	}
	summary := &entity.QuestionSummary{Title: full.Title, Content: body, Answer: answer, QuestionID: full.ID}
	refs := make([]entity.QuestionSummaryReference, 0, len(full.Refers))
	for _, r := range full.Refers {
		refs = append(refs, entity.QuestionSummaryReference{Name: r.Name, URL: r.URL})
	}
	return uc.questionRepo.CreateSummary(ctx, summary, refs)
}

func (uc *QuestionUsecase) upsertSummary(ctx context.Context, q *entity.Question, input QuestionSummaryInput) error {
	full, err := uc.questionRepo.GetByUUID(ctx, q.UUID.String())
	if err != nil {
		return err
	}
	refMap := make(map[string]entity.QuestionRefer, len(full.Refers))
	for _, r := range full.Refers {
		refMap[r.UUID.String()] = r
	}
	refs := make([]entity.QuestionSummaryReference, 0, len(input.ReferUUIDs))
	for _, uid := range input.ReferUUIDs {
		if r, ok := refMap[uid]; ok {
			refs = append(refs, entity.QuestionSummaryReference{
				Name: r.Name,
				URL:  r.URL,
			})
		}
	}
	return uc.questionRepo.UpsertSummary(ctx, full.ID, full.Title, strings.TrimSpace(input.Content), strings.TrimSpace(input.Answer), refs)
}
