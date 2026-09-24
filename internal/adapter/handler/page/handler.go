package page

import (
	"encoding/json"
	"html/template"
	"log/slog"

	"solvi/internal/adapter/handler/authctx"
	loguc "solvi/internal/application/usecase/log_usecase"
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
	Log        *loguc.LogUsecase
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) baseData(c *echo.Context, active string) map[string]interface{} {
	const op = "page.baseData"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	nav := []NavItem{
		{Href: "/", Label: navHomeLabel(claims.IsSupporter || claims.IsAdmin), Icon: "message-circle-question", Active: active == "home"},
		{Href: "/faq", Label: "FAQ", Icon: "book-open", Active: active == "faq"},
		{Href: "/setting", Label: "プロフィール", Icon: "user", Active: active == "setting"},
	}
	if claims.IsSupporter {
		nav = append(nav, NavItem{Href: "/tags", Label: "タグ", Icon: "tags", Active: active == "tags"})
	}
	if claims.IsAdmin {
		nav = append(nav, NavItem{Href: "/management/users", Label: "ユーザ管理", Icon: "users", Active: active == "management"})
		nav = append(nav, NavItem{Href: "/management/logs", Label: "ログ", Icon: "download", Active: active == "management-logs"})
	}
	user, err := h.deps.Setting.GetProfile(ctx, claims.UserID)
	if err != nil {
		logPageWarn(ctx, op, "プロフィール取得失敗", append(pageAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)), slog.String("err", err.Error()))...)
	}
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

func mustJSON(c *echo.Context, v interface{}) template.JS {
	const op = "page.mustJSON"
	b, err := json.Marshal(v)
	if err != nil {
		logPageWarn(requestCtx(c), op, "JSON変換失敗", append(pageAttrs(c), slog.String("err", err.Error()))...)
		return "[]"
	}
	return template.JS(b)
}
