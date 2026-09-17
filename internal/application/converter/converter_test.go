package converter_test

import (
	"strings"
	"testing"
	"time"

	"solvi/internal/application/converter"
	"solvi/internal/domain/entity"
	"solvi/internal/domain/valueobject"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestQuestionEntityToListItem_Branches(t *testing.T) {
	due := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	userUUID := uuid.New()
	qUUID := uuid.New()

	t.Run("with due, content and tags", func(t *testing.T) {
		q := &entity.Question{
			UUID:                  qUUID,
			Title:                 "Q1",
			SupportStatus:         valueobject.SupportStatusSupporting,
			IsRequireHumanSupport: true,
			AnswerDue:             &due,
			Tags:                  []entity.QuestionTag{{Name: "勤怠"}, {Name: "有休"}},
			Contents:              []entity.QuestionContent{{Content: "first content"}, {Content: "second"}},
			QuestionUser: entity.User{
				UUID:           userUUID,
				Name:           "Alice",
				DepartmentName: "Dev",
			},
		}

		out := converter.QuestionEntityToListItem(q)
		if out.UUID != qUUID.String() || out.Title != "Q1" || out.Content != "first content" {
			t.Fatalf("unexpected list item: %+v", out)
		}
		if out.AnswerDue != due.Format(time.RFC3339) {
			t.Fatalf("due mismatch: %s", out.AnswerDue)
		}
		if len(out.Tags) != 2 || out.Tags[0] != "勤怠" {
			t.Fatalf("unexpected tags: %v", out.Tags)
		}
		if out.QuestionUserName != "Alice" || out.SupportStatus != "supporting" {
			t.Fatalf("unexpected fields: %+v", out)
		}
	})

	t.Run("without due and without contents", func(t *testing.T) {
		q := &entity.Question{
			UUID:          qUUID,
			Title:         "Q2",
			SupportStatus: valueobject.SupportStatusPending,
		}
		out := converter.QuestionEntityToListItem(q)
		if out.AnswerDue != "" || out.Content != "" || len(out.Tags) != 0 {
			t.Fatalf("expected empty due/content/tags, got %+v", out)
		}
	})
}

func TestQuestionEntityToDetail_Branches(t *testing.T) {
	due := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)
	qUUID := uuid.New()
	summaryUUID := uuid.New()

	q := &entity.Question{
		Model:                 gorm.Model{ID: 10},
		UUID:                  qUUID,
		Title:                 "Detail Title",
		SupportStatus:         valueobject.SupportStatusDone,
		IsRequireHumanSupport: false,
		AnswerDue:             &due,
		QuestionUserID:        1,
		QuestionUser: entity.User{
			UUID:           uuid.New(),
			Name:           "Bob",
			DepartmentName: "HR",
		},
		Tags: []entity.QuestionTag{{Name: "TagA"}},
		Contents: []entity.QuestionContent{
			{Model: gorm.Model{CreatedAt: now}, UUID: uuid.New(), Content: "Body text", QuestionUser: entity.User{UUID: uuid.New(), Name: "Bob"}},
		},
		Answers: []entity.QuestionAnswer{
			{Model: gorm.Model{CreatedAt: now}, UUID: uuid.New(), Content: "Answer text", AnswerUser: entity.User{UUID: uuid.New(), Name: "Supporter1"}},
		},
		Memos: []entity.QuestionMemo{
			{Model: gorm.Model{CreatedAt: now}, UUID: uuid.New(), Content: "Memo text", MemoUser: entity.User{UUID: uuid.New(), Name: "Supporter2"}},
		},
		Refers: []entity.QuestionRefer{
			{Model: gorm.Model{CreatedAt: now}, UUID: uuid.New(), Name: "Manual", URL: "http://manual", User: entity.User{UUID: uuid.New()}},
		},
		Summary: &entity.QuestionSummary{
			UUID:    summaryUUID,
			Title:   "Summary Title",
			Content: "Summary Content",
			Answer:  "Summary Answer",
			References: []entity.QuestionSummaryReference{
				{UUID: uuid.New(), Name: "Summary Ref", URL: "http://sumref"},
			},
		},
	}

	t.Run("include memos and summary present", func(t *testing.T) {
		out := converter.QuestionEntityToDetail(q, true)
		if len(out.Memos) != 1 || out.Memos[0].Content != "Memo text" {
			t.Fatalf("memos mismatch: %+v", out.Memos)
		}
		if out.Summary == nil || out.Summary.Title != "Summary Title" || len(out.Summary.References) != 1 {
			t.Fatalf("summary mismatch: %+v", out.Summary)
		}
		if len(out.Contents) != 1 || len(out.Answers) != 1 || len(out.Refers) != 1 {
			t.Fatalf("timelines mismatch: %+v", out)
		}
	})

	t.Run("exclude memos and nil summary and nil due", func(t *testing.T) {
		qCopy := *q
		qCopy.AnswerDue = nil
		qCopy.Summary = nil
		out := converter.QuestionEntityToDetail(&qCopy, false)
		if len(out.Memos) != 0 {
			t.Fatalf("expected memos to be omitted, got %+v", out.Memos)
		}
		if out.Summary != nil || out.AnswerDue != "" {
			t.Fatalf("expected nil summary and empty due: %+v", out)
		}
	})
}

func TestUserEntityToOutput_Branches(t *testing.T) {
	uUUID := uuid.New()

	t.Run("without icon", func(t *testing.T) {
		user := &entity.User{
			UUID:           uUUID,
			Name:           "Carol",
			Email:          "carol@example.com",
			DepartmentName: "Dev",
			Icon:           nil,
			IsSupporter:    true,
			IsAdmin:        false,
		}
		out := converter.UserEntityToOutput(user)
		if out.Icon != "" || out.IconBase64 != "" || out.Name != "Carol" || !out.IsSupporter || out.IsAdmin {
			t.Fatalf("unexpected user output: %+v", out)
		}
	})

	testCases := []struct {
		filename   string
		expectedIn string
	}{
		{"avatar.jpg", "data:image/jpeg;base64,"},
		{"avatar.jpeg", "data:image/jpeg;base64,"},
		{"avatar.png", "data:image/png;base64,"},
		{"avatar.webp", "data:image/webp;base64,"},
		{"avatar.svg", "data:image/svg+xml;base64,"},
		{"avatar.bin", "data:application/octet-stream;base64,"},
	}

	for _, tc := range testCases {
		t.Run("icon mime "+tc.filename, func(t *testing.T) {
			fname := tc.filename
			user := &entity.User{
				UUID:     uUUID,
				Name:     "Test",
				Icon:     &fname,
				IconBlob: []byte("sample"),
			}
			out := converter.UserEntityToOutput(user)
			if out.Icon != fname {
				t.Fatalf("expected icon %s, got %s", fname, out.Icon)
			}
			if !strings.HasPrefix(string(out.IconBase64), tc.expectedIn) {
				t.Fatalf("expected prefix %s in %s", tc.expectedIn, out.IconBase64)
			}
		})
	}
}
