package converter

import (
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
	"time"
)

func QuestionEntityToListItem(q *entity.Question) outputmodel.QuestionListItemOutput {
	tags := make([]string, 0, len(q.Tags))
	for _, t := range q.Tags {
		tags = append(tags, t.Name)
	}
	due := ""
	if q.AnswerDue != nil {
		due = q.AnswerDue.Format(time.RFC3339)
	}
	return outputmodel.QuestionListItemOutput{
		UUID:                   q.UUID.String(),
		Title:                  q.Title,
		SupportStatus:          q.SupportStatus.String(),
		IsRequireHumanSupport:  q.IsRequireHumanSupport,
		QuestionUserName:       q.QuestionUser.Name,
		QuestionUserDepartment: q.QuestionUser.DepartmentName,
		AnswerDue:              due,
		Tags:                   tags,
	}
}

func QuestionEntityToDetail(q *entity.Question, includeMemos bool) outputmodel.QuestionDetailOutput {
	tags := make([]string, 0, len(q.Tags))
	for _, t := range q.Tags {
		tags = append(tags, t.Name)
	}
	due := ""
	if q.AnswerDue != nil {
		due = q.AnswerDue.Format(time.RFC3339)
	}
	out := outputmodel.QuestionDetailOutput{
		UUID:                   q.UUID.String(),
		Title:                  q.Title,
		SupportStatus:          q.SupportStatus.String(),
		IsRequireHumanSupport:  q.IsRequireHumanSupport,
		AnswerDue:              due,
		QuestionUserUUID:       q.QuestionUser.UUID.String(),
		QuestionUserID:         q.QuestionUserID,
		QuestionUserName:       q.QuestionUser.Name,
		QuestionUserDepartment: q.QuestionUser.DepartmentName,
		Tags:                   tags,
		Contents:               make([]outputmodel.TimelineOutput, 0, len(q.Contents)),
		Answers:                make([]outputmodel.TimelineOutput, 0, len(q.Answers)),
		Refers:                 make([]outputmodel.ReferOutput, 0, len(q.Refers)),
	}
	for _, c := range q.Contents {
		out.Contents = append(out.Contents, outputmodel.TimelineOutput{
			UUID: c.UUID.String(), Content: c.Content,
			UserUUID: c.QuestionUser.UUID.String(), UserName: c.QuestionUser.Name,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}
	for _, a := range q.Answers {
		out.Answers = append(out.Answers, outputmodel.TimelineOutput{
			UUID: a.UUID.String(), Content: a.Content,
			UserUUID: a.AnswerUser.UUID.String(), UserName: a.AnswerUser.Name,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
		})
	}
	if includeMemos {
		out.Memos = make([]outputmodel.TimelineOutput, 0, len(q.Memos))
		for _, m := range q.Memos {
			out.Memos = append(out.Memos, outputmodel.TimelineOutput{
				UUID: m.UUID.String(), Content: m.Content,
				UserUUID: m.MemoUser.UUID.String(), UserName: m.MemoUser.Name,
				CreatedAt: m.CreatedAt.Format(time.RFC3339),
			})
		}
	}
	for _, r := range q.Refers {
		out.Refers = append(out.Refers, outputmodel.ReferOutput{
			UUID:      r.UUID.String(),
			Name:      r.Name,
			URL:       r.URL,
			UserUUID:  r.User.UUID.String(),
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}
