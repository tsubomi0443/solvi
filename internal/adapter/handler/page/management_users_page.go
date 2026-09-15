package page

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ManagementUsersPage(c *echo.Context) error {
	claims := authctx.Claims(c)
	users, err := h.deps.Management.ListUsers()
	if err != nil {
		return err
	}
	data := h.baseData(c, "management")
	data["Users"] = users
	data["UsersJSON"] = mustJSON(users)
	data["CurrentUserUUID"] = claims.UUID
	return c.Render(http.StatusOK, "management_users.html", data)
}
