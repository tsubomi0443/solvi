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

func (uc *QuestionUsecase) List(actorID uint, isSupporter bool) ([]outputmodel.QuestionListItemOutput, error) {
	if isSupporter {
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

func (uc *QuestionUsecase) Get(actorID uint, isSupporter bool, uuid string) (*outputmodel.QuestionDetailOutput, error) {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if !isSupporter && q.QuestionUserID != actorID {
		return nil, fmt.Errorf("閲覧権限がありません")
	}
	return new(converter.QuestionEntityToDetail(q, isSupporter)), nil
}

func (uc *QuestionUsecase) Create(actorID uint, title, content string, tags []string, requireHuman bool) (*outputmodel.QuestionDetailOutput, error) {
	q := &entity.Question{
		Title:                 title,
		IsRequireHumanSupport: requireHuman,
		SupportStatus:         valueobject.SupportStatusPending,
		QuestionUserID:        actorID,
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
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	return uc.questionRepo.AddRefer(&entity.QuestionRefer{Name: name, URL: url, QuestionID: q.ID, UserID: actorID})
}

func (uc *QuestionUsecase) AddMemo(actorID uint, uuid, content string) error {
	q, err := uc.questionRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	return uc.questionRepo.AddMemo(&entity.QuestionMemo{Content: content, QuestionID: q.ID, MemoUserID: actorID})
}

func (uc *QuestionUsecase) Update(actorID uint, isSupporter bool, uuid string, status *string, due *time.Time, tags []string, requireHuman *bool, complete bool) error {
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
		return fmt.Errorf("権限がありません")
	}
	if status != nil {
		parsed, err := valueobject.ParseSupportStatus(supportStatusToInt(*status))
		if err != nil {
			return err
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
		tagEntities := make([]entity.QuestionTag, 0, len(tags))
		for _, t := range tags {
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
