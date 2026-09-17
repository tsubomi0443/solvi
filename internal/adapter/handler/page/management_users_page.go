package page

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ManagementUsersPage(c *echo.Context) error {
	const op = "page.ManagementUsersPage"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	users, err := h.deps.Management.ListUsers(ctx)
	if err != nil {
		logPageError(ctx, op, "ユーザ一覧取得失敗", err, pageAttrs(c)...)
		return err
	}
	data := h.baseData(c, "management")
	data["Users"] = users
	data["UsersJSON"] = mustJSON(c, users)
	data["CurrentUserUUID"] = claims.UUID
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "management_users.html", data)
}
