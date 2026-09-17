package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"solvi/internal/adapter/handler/authctx"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/shared/datetime"

	"github.com/labstack/echo/v5"
)

func (h *Handler) GetQuestion(c *echo.Context) error {
	const op = "api.GetQuestion"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	detail, err := h.deps.Question.Get(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "質問取得失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) ListQuestions(c *echo.Context) error {
	const op = "api.ListQuestions"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	items, err := h.deps.Question.List(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin)
	if err != nil {
		logHandlerDebug(ctx, op, "質問一覧取得失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateQuestion(c *echo.Context) error {
	const op = "api.CreateQuestion"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Title                 string   `json:"title"`
		Content               string   `json:"content"`
		Tags                  []string `json:"tags"`
		AnswerDueText         *string  `json:"answerDue"`
		IsRequireHumanSupport bool     `json:"isRequireHumanSupport"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	logHandlerDebug(ctx, op, "リクエスト受信", http.StatusOK, append(handlerAttrs(c),
		slog.Int("content_len", len(req.Content)),
		slog.Int("tag_count", len(req.Tags)),
		slog.Bool("is_require_human_support", req.IsRequireHumanSupport),
	)...)
	var answerDue *time.Time
	if req.AnswerDueText != nil && strings.TrimSpace(*req.AnswerDueText) != "" {
		parsed, err := datetime.ParseAnswerDueDate(strings.TrimSpace(*req.AnswerDueText))
		if err != nil {
			logHandlerWarn(ctx, op, "回答期限形式不正", http.StatusBadRequest, handlerAttrs(c)...)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid due format"})
		}
		answerDue = &parsed
	}
	detail, err := h.deps.Question.Create(ctx, claims.UserID, req.Title, req.Content, req.Tags, answerDue, req.IsRequireHumanSupport)
	if err != nil {
		logHandlerDebug(ctx, op, "質問作成失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendToQuestion("create-question", detail, claims.UserID)
	return c.JSON(http.StatusCreated, detail)
}

func (h *Handler) AppendContent(c *echo.Context) error {
	const op = "api.AppendContent"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AppendContent(ctx, claims.UserID, uuid, req.Content); err != nil {
		logHandlerDebug(ctx, op, "追記失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	h.deps.Hub.SendToQuestion("create-content", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) AddAnswer(c *echo.Context) error {
	const op = "api.AddAnswer"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Content string                    `json:"content"`
		Refers  []outputmodel.ReferOutput `json:"refers"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddAnswer(ctx, claims.UserID, uuid, req.Content, req.Refers); err != nil {
		logHandlerDebug(ctx, op, "回答追加失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	h.deps.Hub.SendToQuestion("create-answer", q, q.QuestionUserID)
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) AddMemo(c *echo.Context) error {
	const op = "api.AddMemo"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddMemo(ctx, claims.UserID, uuid, req.Content); err != nil {
		logHandlerDebug(ctx, op, "メモ追加失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	h.deps.Hub.SendMemo("create-memo", q)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) AddRefer(c *echo.Context) error {
	const op = "api.AddRefer"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Refers []outputmodel.ReferOutput `json:"refers"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddRefers(ctx, claims.UserID, uuid, req.Refers); err != nil {
		msg := err.Error()
		if msg == "タイトルとURLは両方入力してください" || msg == "引用情報がありません" {
			logHandlerWarn(ctx, op, "引用入力不正", http.StatusBadRequest, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": msg})
		}
		logHandlerDebug(ctx, op, "引用追加失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": msg})
	}
	q, _ := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	h.deps.Hub.SendToQuestion("create-refer", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) UpdateQuestion(c *echo.Context) error {
	const op = "api.UpdateQuestion"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Title                 *string   `json:"title"`
		Status                *string   `json:"status"`
		AnswerDue             *string   `json:"answerDue"`
		Tags                  *[]string `json:"tags"`
		IsRequireHumanSupport *bool     `json:"isRequireHumanSupport"`
		Complete              bool      `json:"complete"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	var due *time.Time
	if req.AnswerDue != nil && strings.TrimSpace(*req.AnswerDue) != "" {
		parsed, err := datetime.ParseAnswerDueDate(strings.TrimSpace(*req.AnswerDue))
		if err != nil {
			logHandlerWarn(ctx, op, "回答期限形式不正", http.StatusBadRequest, handlerAttrs(c)...)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date"})
		}
		due = &parsed
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.Update(ctx, claims.UserID, claims.IsAdmin, claims.IsSupporter, uuid, req.Title, req.Status, due, req.Tags, req.IsRequireHumanSupport, req.Complete); err != nil {
		logHandlerDebug(ctx, op, "質問更新失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, q)
}

func (h *Handler) DeleteAnswer(c *echo.Context) error {
	const op = "api.DeleteAnswer"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	answerUUID := c.Param("answerUuid")
	if err := h.deps.Question.DeleteAnswer(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid, answerUUID); err != nil {
		logHandlerDebug(ctx, op, "回答削除失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid), slog.String("answer_uuid", answerUUID))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, err := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "削除後の質問取得失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteMemo(c *echo.Context) error {
	const op = "api.DeleteMemo"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	memoUUID := c.Param("memoUuid")
	if err := h.deps.Question.DeleteMemo(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid, memoUUID); err != nil {
		logHandlerDebug(ctx, op, "メモ削除失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid), slog.String("memo_uuid", memoUUID))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, err := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "削除後の質問取得失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendMemo("update-question", q)
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteRefer(c *echo.Context) error {
	const op = "api.DeleteRefer"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	referUUID := c.Param("referUuid")
	if err := h.deps.Question.DeleteRefer(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid, referUUID); err != nil {
		logHandlerDebug(ctx, op, "引用削除失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid), slog.String("refer_uuid", referUUID))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, err := h.deps.Question.Get(ctx, claims.UserID, true, claims.IsAdmin, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "削除後の質問取得失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteQuestion(c *echo.Context) error {
	const op = "api.DeleteQuestion"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	if err := h.deps.Question.Delete(ctx, claims.IsAdmin, uuid); err != nil {
		logHandlerDebug(ctx, op, "質問削除失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	question, err := h.deps.Question.Get(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "削除後の質問取得失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.String("question_uuid", uuid))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendToQuestion("delete-question", uuid, question.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}
