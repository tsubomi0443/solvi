package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ManagementUsersPage(c *echo.Context) error {
	users, err := h.deps.Management.ListUsers()
	if err != nil {
		return err
	}
	data := h.baseData(c, "management")
	data["Users"] = users
	data["UsersJSON"] = mustJSON(users)
	return c.Render(http.StatusOK, "management_users.html", data)
}
