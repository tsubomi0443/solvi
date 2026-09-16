package api

import (
	"fmt"
	"net/http"
	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) IssueLogDownloadAll(c *echo.Context) error {
	claims := authctx.Claims(c)
	result, err := h.deps.Log.IssueAll(claims.UUID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) IssueLogDownloadDate(c *echo.Context) error {
	claims := authctx.Claims(c)
	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "date is required"})
	}

	result, err := h.deps.Log.IssueDate(claims.UUID, date)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) StreamLogDownload(c *echo.Context) error {
	claims := authctx.Claims(c)
	key := c.Param("key")

	ticket, err := h.deps.Log.LookupTicket(key, claims.UUID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	res := c.Response()
	res.Header().Set("Content-Type", "application/zip")
	res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", ticket.Filename))

	if err := h.deps.Log.Stream(res, key, claims.UUID); err != nil {
		return err
	}
	return nil
}
