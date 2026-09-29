package lineworks_usecase

import (
	"context"
	"log/slog"
	"time"

	"solvi/internal/domain/entity"
	"solvi/internal/domain/entity/lineworks"
	"solvi/internal/domain/valueobject"
	"solvi/internal/shared/config"
	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
)

// DueDigestSchedule は定期通知の実行条件。
type DueDigestSchedule struct {
	NoticeTimes []config.NoticeTime
	ChannelIDs  []string
	AppBaseURL  string
}

func (uc *Usecase) EnqueueDueDigestIfScheduled(ctx context.Context, now time.Time) error {
	const op = opLineWorks + ".EnqueueDueDigestIfScheduled"
	if len(uc.schedule.NoticeTimes) == 0 || len(uc.schedule.ChannelIDs) == 0 {
		return nil
	}
	loc := uc.jst
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}
	}
	matched, ok := matchNoticeTime(uc.schedule.NoticeTimes, now, loc)
	if !ok {
		return nil
	}

	inJST := now.In(loc)
	today := dateOnly(inJST)
	tomorrow := today.AddDate(0, 0, 1)
	questions, err := uc.questionRepo.ListIncompleteDueOnDates(ctx, today, tomorrow)
	if err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "期日通知対象の取得失敗", slog.String("err", err.Error()))
		return err
	}

	todayItems, tomorrowItems := splitDueDigestItems(questions, uc.schedule.AppBaseURL, today, tomorrow)
	body := DueDigestMessage(todayItems, tomorrowItems)
	statusRevision := statusRevisionFromDate(today)
	notices := make([]lineworks.Notification, 0, len(uc.schedule.ChannelIDs))
	for _, chID := range uc.schedule.ChannelIDs {
		notices = append(notices, lineworks.Notification{
			QuestionUUID:   uuid.Nil,
			Event:          valueobject.LineWorksEventDueDigest,
			StatusRevision: statusRevision,
			Burst:          matched.Burst(),
			ChannelID:      chID,
			Destination:    valueobject.LineWorksDestinationChannel,
			Body:           body,
			Status:         valueobject.LineWorksJobPending,
			NextAttemptAt:  now,
		})
	}
	if err := uc.repo.EnqueueDueDigests(ctx, notices); err != nil {
		logutils.Error(ctx, logutils.LayerUsecase, op, "期日通知の登録失敗", slog.String("err", err.Error()))
		return err
	}
	return nil
}

func matchNoticeTime(times []config.NoticeTime, now time.Time, loc *time.Location) (config.NoticeTime, bool) {
	for _, t := range times {
		if t.Matches(now, loc) {
			return t, true
		}
	}
	return config.NoticeTime{}, false
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func statusRevisionFromDate(day time.Time) int {
	y, m, d := day.Date()
	return y*10000 + int(m)*100 + d
}

func splitDueDigestItems(questions []entity.Question, appBaseURL string, today, tomorrow time.Time) ([]DueDigestItem, []DueDigestItem) {
	todayItems := make([]DueDigestItem, 0)
	tomorrowItems := make([]DueDigestItem, 0)
	for _, q := range questions {
		if q.AnswerDue == nil {
			continue
		}
		dueDay := dateOnly(q.AnswerDue.In(today.Location()))
		item := DueDigestItem{
			Title:     q.Title,
			AskerName: q.QuestionUser.Name,
			DueDate:   dueDay.Format("2006-01-02"),
			Link:      QuestionLink(appBaseURL, q.UUID.String()),
		}
		switch {
		case dueDay.Equal(today):
			todayItems = append(todayItems, item)
		case dueDay.Equal(tomorrow):
			tomorrowItems = append(tomorrowItems, item)
		}
	}
	return todayItems, tomorrowItems
}
