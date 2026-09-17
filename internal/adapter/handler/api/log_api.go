package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) IssueLogDownloadAll(c *echo.Context) error {
	const op = "api.IssueLogDownloadAll"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	result, err := h.deps.Log.IssueAll(ctx, claims.UUID)
	if err != nil {
		logHandlerDebug(ctx, op, "ログダウンロード発行失敗", http.StatusNotFound, append(handlerAttrs(c), slog.String("user_uuid", claims.UUID))...)
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) IssueLogDownloadDate(c *echo.Context) error {
	const op = "api.IssueLogDownloadDate"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	date := c.QueryParam("date")
	if date == "" {
		logHandlerWarn(ctx, op, "日付未指定", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "date is required"})
	}

	result, err := h.deps.Log.IssueDate(ctx, claims.UUID, date)
	if err != nil {
		logHandlerDebug(ctx, op, "ログダウンロード発行失敗", http.StatusNotFound, append(handlerAttrs(c), slog.String("user_uuid", claims.UUID), slog.String("date", date))...)
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) StreamLogDownload(c *echo.Context) error {
	const op = "api.StreamLogDownload"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	key := c.Param("key")

	ticket, err := h.deps.Log.LookupTicket(ctx, key, claims.UUID)
	if err != nil {
		logHandlerDebug(ctx, op, "チケット照会失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("key", key), slog.String("user_uuid", claims.UUID))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	res := c.Response()
	res.Header().Set("Content-Type", "application/zip")
	res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", ticket.Filename))

	logHandlerInfo(ctx, op, "ログストリーム開始", slog.String("key", key), slog.String("user_uuid", claims.UUID), slog.String("filename", ticket.Filename))
	if err := h.deps.Log.Stream(ctx, res, key, claims.UUID); err != nil {
		logHandlerError(ctx, op, "ログストリーム失敗", err, append(handlerAttrs(c), slog.String("key", key), slog.String("user_uuid", claims.UUID))...)
		return err
	}
	logHandlerInfo(ctx, op, "ログストリーム完了", slog.String("key", key), slog.String("user_uuid", claims.UUID), slog.String("filename", ticket.Filename))
	return nil
}
