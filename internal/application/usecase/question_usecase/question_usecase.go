package question_usecase

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"solvi/internal/application/converter"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
	ext "solvi/internal/domain/interface/external"
	repo "solvi/internal/domain/interface/repository"
	"solvi/internal/domain/valueobject"
	"solvi/internal/shared/config"
)

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

func (uc *QuestionUsecase) List(actorID uint, isSupporter, isAdmin bool) ([]outputmodel.QuestionListItemOutput, error) {
	if canViewAllQuestions(isSupporter, isAdmin) {
		qs, err := uc.questionRepo.ListAll()
		if err != nil {
			return nil, err
		}
		return mapList(qs), nil
	}
	qs, err := uc.questionRepo.ListByQuestionUserID(actorID)
	if err != nil {
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

func (uc *QuestionUsecase) Get(actorID uint, isSupporter, isAdmin bool, uuid string) (*outputmodel.QuestionDetailOutput, error) {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if !canViewAllQuestions(isSupporter, isAdmin) && q.QuestionUserID != actorID {
		return nil, fmt.Errorf("閲覧権限がありません")
	}
	return new(converter.QuestionEntityToDetail(q, canViewAllQuestions(isSupporter, isAdmin))), nil
}

func (uc *QuestionUsecase) Create(actorID uint, title, content string, tags []string, answerDue *time.Time, requireHuman bool) (*outputmodel.QuestionDetailOutput, error) {
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
	if err := uc.questionRepo.Create(q); err != nil {
		return nil, err
	}
	created, err := uc.questionRepo.GetByUUID(q.UUID.String())
	if err != nil {
		return nil, err
	}
	// TODO: AI回答機能は未実装の状態として進める。そのため一旦コメントアウトで対応しています。
	//// go uc.runAI(created.UUID.String(), created.Title, content)
	return new(converter.QuestionEntityToDetail(created, false)), nil
}

func (uc *QuestionUsecase) runAI(questionUUID, title, content string) {
	result, err := uc.bedrock.AnswerQuestion(title, content)
	if err != nil {
		slog.Error("AI回答失敗", "err", err, "uuid", questionUUID)
		q, e := uc.questionRepo.GetByUUID(questionUUID)
		if e == nil {
			q.IsRequireHumanSupport = true
			_ = uc.questionRepo.Update(q)
		}
		return
	}
	sys, err := uc.userRepo.GetByEmail(config.GetSystemUserEmail())
	if err != nil {
		slog.Error("system user not found", "err", err)
		return
	}
	q, err := uc.questionRepo.GetByUUID(questionUUID)
	if err != nil {
		return
	}
	_ = uc.questionRepo.AddAnswer(&entity.QuestionAnswer{Content: result.Content, AnswerUserID: sys.ID, QuestionID: q.ID})
	for _, ref := range result.References {
		_ = uc.questionRepo.AddRefer(&entity.QuestionRefer{Name: ref.Name, URL: ref.URL, QuestionID: q.ID, UserID: sys.ID})
	}
	if uc.onAIAnswer != nil {
		uc.onAIAnswer(questionUUID)
	}
}

func (uc *QuestionUsecase) AppendContent(actorID uint, uuid, content string) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	if q.QuestionUserID != actorID {
		return fmt.Errorf("権限がありません")
	}
	return uc.questionRepo.AddContent(&entity.QuestionContent{Content: content, QuestionUserID: actorID, QuestionID: q.ID})
}

func (uc *QuestionUsecase) AddAnswer(actorID uint, uuid, content string, refers []outputmodel.ReferOutput) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	if err := uc.questionRepo.AddAnswer(&entity.QuestionAnswer{Content: content, AnswerUserID: actorID, QuestionID: q.ID}); err != nil {
		return err
	}
	for _, r := range refers {
		_ = uc.questionRepo.AddRefer(&entity.QuestionRefer{Name: r.Name, URL: r.URL, QuestionID: q.ID, UserID: actorID})
	}
	if q.SupportStatus == valueobject.SupportStatusPending {
		q.SupportStatus = valueobject.SupportStatusSupporting
		_ = uc.questionRepo.Update(q)
	}
	return nil
}

func (uc *QuestionUsecase) AddRefer(actorID uint, uuid, name, url string) error {
	return uc.AddRefers(actorID, uuid, []outputmodel.ReferOutput{{Name: name, URL: url}})
}

func (uc *QuestionUsecase) AddRefers(actorID uint, uuid string, refers []outputmodel.ReferOutput) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
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
			return fmt.Errorf("タイトルとURLは両方入力してください")
		}
		valid = append(valid, outputmodel.ReferOutput{Name: name, URL: url})
	}
	if len(valid) == 0 {
		return fmt.Errorf("引用情報がありません")
	}
	for _, r := range valid {
		if err := uc.questionRepo.AddRefer(&entity.QuestionRefer{
			Name: r.Name, URL: r.URL, QuestionID: q.ID, UserID: actorID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (uc *QuestionUsecase) AddMemo(actorID uint, uuid, content string) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	return uc.questionRepo.AddMemo(&entity.QuestionMemo{Content: content, QuestionID: q.ID, MemoUserID: actorID})
}

func (uc *QuestionUsecase) DeleteAnswer(actorID uint, isSupporter, isAdmin bool, questionUUID, answerUUID string) error {
	if !isSupporter && !isAdmin {
		return fmt.Errorf("権限がありません")
	}
	q, err := uc.questionRepo.GetByUUID(questionUUID)
	if err != nil {
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
		return fmt.Errorf("回答が見つかりません")
	}
	if !isAdmin && target.AnswerUserID != actorID {
		return fmt.Errorf("権限がありません")
	}
	return uc.questionRepo.SoftDeleteAnswerByUUID(answerUUID)
}

func (uc *QuestionUsecase) DeleteMemo(actorID uint, isSupporter, isAdmin bool, questionUUID, memoUUID string) error {
	if !isSupporter && !isAdmin {
		return fmt.Errorf("権限がありません")
	}
	q, err := uc.questionRepo.GetByUUID(questionUUID)
	if err != nil {
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
		return fmt.Errorf("メモが見つかりません")
	}
	if !isAdmin && target.MemoUserID != actorID {
		return fmt.Errorf("権限がありません")
	}
	return uc.questionRepo.SoftDeleteMemoByUUID(memoUUID)
}

func (uc *QuestionUsecase) Delete(isAdmin bool, uuid string) error {
	if !isAdmin {
		return fmt.Errorf("権限がありません")
	}
	if _, err := uc.questionRepo.GetByUUID(uuid); err != nil {
		return err
	}
	return uc.questionRepo.SoftDeleteByUUID(uuid)
}

func (uc *QuestionUsecase) Update(actorID uint, isSupporter bool, uuid string, title *string, status *string, due *time.Time, tags *[]string, requireHuman *bool, complete bool) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	if complete {
		if isSupporter || (!q.IsRequireHumanSupport && q.QuestionUserID == actorID) {
			q.SupportStatus = valueobject.SupportStatusDone
			if err := uc.questionRepo.Update(q); err != nil {
				return err
			}
			return uc.createSummary(q)
		}
		return fmt.Errorf("完了できません")
	}
	if !isSupporter {
		if q.QuestionUserID != actorID {
			return fmt.Errorf("権限がありません")
		}
		if title != nil || status != nil || due != nil || tags != nil {
			return fmt.Errorf("権限がありません")
		}
		if requireHuman == nil {
			return fmt.Errorf("権限がありません")
		}
		q.IsRequireHumanSupport = *requireHuman
		return uc.questionRepo.Update(q)
	}
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return fmt.Errorf("タイトルは必須です")
		}
		q.Title = trimmed
	}
	if status != nil {
		parsed, err := valueobject.ParseSupportStatus(supportStatusToInt(*status))
		if err != nil {
			return err
		}
		if parsed == valueobject.SupportStatusDone && q.SupportStatus != valueobject.SupportStatusDone {
			q.SupportStatus = parsed
			if err := uc.questionRepo.Update(q); err != nil {
				return err
			}
			return uc.createSummary(q)
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
		if err := uc.questionRepo.ReplaceTags(q.ID, tagEntities); err != nil {
			return err
		}
	}
	return uc.questionRepo.Update(q)
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

func (uc *QuestionUsecase) createSummary(q *entity.Question) error {
	full, err := uc.questionRepo.GetByUUID(q.UUID.String())
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
	return uc.questionRepo.CreateSummary(summary, refs)
}
