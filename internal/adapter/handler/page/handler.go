package page

import (
	"encoding/json"
	"html/template"

	"solvi/internal/adapter/handler/authctx"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"

	"github.com/labstack/echo/v5"
)

type NavItem struct {
	Href   string
	Label  string
	Icon   string
	Active bool
}

type Deps struct {
	Question   *quc.QuestionUsecase
	Setting    *setuc.SettingUsecase
	Management *mnguc.ManagementUsecase
	Tag        *taguc.TagUsecase
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) baseData(c *echo.Context, active string) map[string]interface{} {
	claims := authctx.Claims(c)
	nav := []NavItem{
		{Href: "/", Label: navHomeLabel(claims.IsSupporter || claims.IsAdmin), Icon: "message-circle-question", Active: active == "home"},
		{Href: "/setting", Label: "プロフィール", Icon: "user", Active: active == "setting"},
	}
	if claims.IsSupporter {
		nav = append(nav, NavItem{Href: "/tags", Label: "タグ", Icon: "tags", Active: active == "tags"})
	}
	if claims.IsAdmin {
		nav = append(nav, NavItem{Href: "/management/users", Label: "ユーザ管理", Icon: "users", Active: active == "management"})
	}
	user, _ := h.deps.Setting.GetProfile(claims.UserID)
	return map[string]interface{}{
		"Navigation":  nav,
		"User":        user,
		"IsSupporter": claims.IsSupporter,
		"IsAdmin":     claims.IsAdmin,
		"CanViewAll":  claims.IsSupporter || claims.IsAdmin,
	}
}

func navHomeLabel(isSupporter bool) string {
	if isSupporter {
		return "質問一覧"
	}
	return "自分の質問"
}

func mustJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return template.JS(b)
}
