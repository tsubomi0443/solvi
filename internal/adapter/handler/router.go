package handler

import (
	"log/slog"

	"solvi/internal/adapter/handler/api"
	"solvi/internal/adapter/handler/authctx"
	"solvi/internal/adapter/handler/page"
	"solvi/internal/adapter/handler/sse"
	authuc "solvi/internal/application/usecase/auth_usecase"
	loguc "solvi/internal/application/usecase/log_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"
	logutils "solvi/internal/shared/logUtils"

	"github.com/labstack/echo/v5"
)

type Deps struct {
	Auth       *authuc.AuthUsecase
	Question   *quc.QuestionUsecase
	Setting    *setuc.SettingUsecase
	Management *mnguc.ManagementUsecase
	Tag        *taguc.TagUsecase
	Log        *loguc.LogUsecase
	Hub        *sse.Hub
}

func RegisterRoutes(e *echo.Echo, deps Deps, auditLogger *slog.Logger) {
	e.Static("/static", "static")
	e.Static("/uploads", "uploads")
	e.File("/favicon.ico", "favicon.ico")

	ph := page.New(page.Deps{
		Question:   deps.Question,
		Setting:    deps.Setting,
		Management: deps.Management,
		Tag:        deps.Tag,
		Log:        deps.Log,
	})
	ah := api.New(api.Deps{
		Auth:       deps.Auth,
		Question:   deps.Question,
		Setting:    deps.Setting,
		Management: deps.Management,
		Tag:        deps.Tag,
		Log:        deps.Log,
		Hub:        deps.Hub,
	})
	sh := sse.NewHandler(deps.Hub)

	e.GET("/login", ph.LoginPage)
	e.POST("/api/v1/login", ah.Login, auditLogging(auditLogger, true))

	auth := e.Group("", JWTConfig(auditLogger))
	auth.GET("/", ph.HomePage, auditLogging(auditLogger, false))
	auth.GET("/questions/new", ph.QuestionNewPage, auditLogging(auditLogger, false))
	auth.GET("/questions/:uuid", ph.QuestionDetailPage, auditLogging(auditLogger, false))
	auth.GET("/setting", ph.SettingPage, auditLogging(auditLogger, false))
	auth.GET("/tags", ph.TagsPage, SupporterOnlyPage, auditLogging(auditLogger, false))
	auth.GET("/management/users", ph.ManagementUsersPage, AdminOnly, auditLogging(auditLogger, false))
	auth.GET("/management/logs", ph.ManagementLogsPage, AdminOnly, auditLogging(auditLogger, false))
	auth.GET("/sse", sh.Stream, auditLogging(auditLogger, false))
	auth.GET("/logout", ph.LogoutPage, auditLogging(auditLogger, true))

	apiAuth := e.Group("/api/v1", APIJWTConfig(auditLogger))
	apiAuth.GET("/user/icon/:uuid", ah.GetUserIcon, auditLogging(auditLogger, false))
	apiAuth.GET("/questions", ah.ListQuestions, auditLogging(auditLogger, false))
	apiAuth.GET("/questions/:uuid", ah.GetQuestion, auditLogging(auditLogger, false))
	apiAuth.POST("/questions", ah.CreateQuestion, auditLogging(auditLogger, false))
	apiAuth.POST("/questions/:uuid/contents", ah.AppendContent, auditLogging(auditLogger, false))
	apiAuth.POST("/questions/:uuid/answers", ah.AddAnswer, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.DELETE("/questions/:uuid", ah.DeleteQuestion, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.DELETE("/questions/:uuid/answers/:answerUuid", ah.DeleteAnswer, SupporterOrAdmin, auditLogging(auditLogger, false))
	apiAuth.POST("/questions/:uuid/memos", ah.AddMemo, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.DELETE("/questions/:uuid/memos/:memoUuid", ah.DeleteMemo, SupporterOrAdmin, auditLogging(auditLogger, false))
	apiAuth.POST("/questions/:uuid/refers", ah.AddRefer, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.DELETE("/questions/:uuid/refers/:referUuid", ah.DeleteRefer, SupporterOrAdmin, auditLogging(auditLogger, false))
	apiAuth.PUT("/questions/:uuid", ah.UpdateQuestion, auditLogging(auditLogger, false))
	apiAuth.PUT("/setting", ah.UpdateSetting, auditLogging(auditLogger, false))
	apiAuth.POST("/setting/icon", ah.UploadIcon, auditLogging(auditLogger, false))
	apiAuth.DELETE("/setting/icon", ah.DeleteIcon, auditLogging(auditLogger, false))
	apiAuth.GET("/management/users", ah.ListUsers, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.PUT("/management/users/:uuid", ah.UpdateUser, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.DELETE("/management/users/:uuid", ah.DeleteUser, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.GET("/tags", ah.ListTags, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.PUT("/tags", ah.RenameTag, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.DELETE("/tags", ah.DeleteTag, SupporterOnly, auditLogging(auditLogger, false))
	apiAuth.POST("/log/download/all", ah.IssueLogDownloadAll, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.POST("/log/download/date", ah.IssueLogDownloadDate, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.GET("/log/download/:key", ah.StreamLogDownload, AdminOnlyAPI, auditLogging(auditLogger, false))
	apiAuth.POST("/api/v1/logout", ah.Logout, auditLogging(auditLogger, true))
}

func auditLogging(auditLogger *slog.Logger, beforeAuth bool) echo.MiddlewareFunc {
	if auditLogger == nil {
		auditLogger = slog.Default()
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			attrs := []any{
				slog.String("request_id", logutils.RequestIDFromContext(req.Context())),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("remote_addr", req.RemoteAddr),
				slog.String("real_ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
				slog.String("uuid", c.Param("uuid")),
				slog.String("key", c.Param("key")),
			}
			if !beforeAuth {
				attrs = append(attrs,
					logutils.LogAny("query", c.QueryParams()),
					logutils.LogAny("claims", authctx.Claims(c)),
				)
			}
			auditLogger.Info("audit", attrs...)
			return next(c)
		}
	}
}
