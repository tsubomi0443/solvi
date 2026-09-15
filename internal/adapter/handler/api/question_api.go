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
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	detail, err := h.deps.Question.Get(claims.UserID, claims.IsSupporter, uuid)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, detail)

}

func (h *Handler) ListQuestions(c *echo.Context) error {
	claims := authctx.Claims(c)
	items, err := h.deps.Question.List(claims.UserID, claims.IsSupporter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)

}

func (h *Handler) CreateQuestion(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Title                 string   `json:"title"`
		Content               string   `json:"content"`
		Tags                  []string `json:"tags"`
		AnswerDueText         *string  `json:"answerDue"`
		IsRequireHumanSupport bool     `json:"isRequireHumanSupport"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	slog.Info("Requestのバインド結果中身", slog.Any("data", req))
	var answerDue *time.Time
	if req.AnswerDueText != nil && strings.TrimSpace(*req.AnswerDueText) != "" {
		parsed, err := datetime.ParseAnswerDueDate(strings.TrimSpace(*req.AnswerDueText))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid due format"})
		}
		answerDue = &parsed
	}
	detail, err := h.deps.Question.Create(claims.UserID, req.Title, req.Content, req.Tags, answerDue, req.IsRequireHumanSupport)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	h.deps.Hub.SendToSupporters("create-question", detail)
	return c.JSON(http.StatusCreated, detail)

}

func (h *Handler) AppendContent(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AppendContent(claims.UserID, uuid, req.Content); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(claims.UserID, claims.IsSupporter, uuid)
	h.deps.Hub.SendToQuestion("create-content", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})

}

func (h *Handler) AddAnswer(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Content string                    `json:"content"`
		Refers  []outputmodel.ReferOutput `json:"refers"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddAnswer(claims.UserID, uuid, req.Content, req.Refers); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(claims.UserID, true, uuid)
	h.deps.Hub.SendToQuestion("create-answer", q, q.QuestionUserID)
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})

}

func (h *Handler) AddMemo(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddMemo(claims.UserID, uuid, req.Content); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(claims.UserID, true, uuid)
	h.deps.Hub.SendMemo("create-memo", q)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})

}

func (h *Handler) AddRefer(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req outputmodel.ReferOutput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.AddRefer(claims.UserID, uuid, req.Name, req.URL); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(claims.UserID, true, uuid)
	h.deps.Hub.SendToQuestion("create-refer", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})

}

func (h *Handler) UpdateQuestion(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Title                 *string  `json:"title"`
		Status                *string  `json:"status"`
		AnswerDue             *string  `json:"answerDue"`
		Tags                  []string `json:"tags"`
		IsRequireHumanSupport *bool    `json:"isRequireHumanSupport"`
		Complete              bool     `json:"complete"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	var due *time.Time
	if req.AnswerDue != nil && strings.TrimSpace(*req.AnswerDue) != "" {
		parsed, err := datetime.ParseAnswerDueDate(strings.TrimSpace(*req.AnswerDue))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date"})
		}
		due = &parsed
	}
	uuid := c.Param("uuid")
	if err := h.deps.Question.Update(claims.UserID, claims.IsSupporter, uuid, req.Title, req.Status, due, req.Tags, req.IsRequireHumanSupport, req.Complete); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	q, _ := h.deps.Question.Get(claims.UserID, claims.IsSupporter, uuid)
	h.deps.Hub.SendToQuestion("update-question", q, q.QuestionUserID)
	return c.JSON(http.StatusOK, q)

}
