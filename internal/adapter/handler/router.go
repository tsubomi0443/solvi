package handler

import (
	"io"
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

func RegisterRoutes(e *echo.Echo, deps Deps, accessLog io.Writer) {
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
	e.GET("/logout", ph.LogoutPage)

	auth := e.Group("", JWTConfig())
	auth.GET("/", ph.HomePage, infoLogging)
	auth.GET("/questions/new", ph.QuestionNewPage, infoLogging)
	auth.GET("/questions/:uuid", ph.QuestionDetailPage, infoLogging)
	auth.GET("/setting", ph.SettingPage, infoLogging)
	auth.GET("/tags", ph.TagsPage, SupporterOnlyPage, infoLogging)
	auth.GET("/management/users", ph.ManagementUsersPage, AdminOnly, infoLogging)
	auth.GET("/management/logs", ph.ManagementLogsPage, AdminOnly, infoLogging)
	auth.GET("/sse", sh.Stream, infoLogging)

	apiAuth := e.Group("/api/v1", APIJWTConfig())
	apiAuth.POST("/login", ah.Login, authenticateInfoLogging, infoLogging)
	apiAuth.POST("/logout", ah.Logout, authenticateInfoLogging, infoLogging)
	apiAuth.GET("/user/icon/:uuid", ah.GetUserIcon, infoLogging)
	apiAuth.GET("/questions", ah.ListQuestions, infoLogging)
	apiAuth.GET("/questions/:uuid", ah.GetQuestion, infoLogging)
	apiAuth.POST("/questions", ah.CreateQuestion, infoLogging)
	apiAuth.POST("/questions/:uuid/contents", ah.AppendContent, infoLogging)
	apiAuth.POST("/questions/:uuid/answers", ah.AddAnswer, SupporterOnly, infoLogging)
	apiAuth.DELETE("/questions/:uuid", ah.DeleteQuestion, AdminOnlyAPI, infoLogging)
	apiAuth.DELETE("/questions/:uuid/answers/:answerUuid", ah.DeleteAnswer, SupporterOrAdmin, infoLogging)
	apiAuth.POST("/questions/:uuid/memos", ah.AddMemo, SupporterOnly, infoLogging)
	apiAuth.DELETE("/questions/:uuid/memos/:memoUuid", ah.DeleteMemo, SupporterOrAdmin, infoLogging)
	apiAuth.POST("/questions/:uuid/refers", ah.AddRefer, SupporterOnly, infoLogging)
	apiAuth.DELETE("/questions/:uuid/refers/:referUuid", ah.DeleteRefer, SupporterOrAdmin, infoLogging)
	apiAuth.PUT("/questions/:uuid", ah.UpdateQuestion, infoLogging)
	apiAuth.PUT("/setting", ah.UpdateSetting, infoLogging)
	apiAuth.POST("/setting/icon", ah.UploadIcon, infoLogging)
	apiAuth.DELETE("/setting/icon", ah.DeleteIcon, infoLogging)
	apiAuth.GET("/management/users", ah.ListUsers, AdminOnlyAPI, infoLogging)
	apiAuth.PUT("/management/users/:uuid", ah.UpdateUser, AdminOnlyAPI, infoLogging)
	apiAuth.GET("/tags", ah.ListTags, SupporterOnly, infoLogging)
	apiAuth.PUT("/tags", ah.RenameTag, SupporterOnly, infoLogging)
	apiAuth.DELETE("/tags", ah.DeleteTag, SupporterOnly, infoLogging)
	apiAuth.POST("/log/download/all", ah.IssueLogDownloadAll, AdminOnlyAPI, infoLogging)
	apiAuth.POST("/log/download/date", ah.IssueLogDownloadDate, AdminOnlyAPI, infoLogging)
	apiAuth.GET("/log/download/:key", ah.StreamLogDownload, AdminOnlyAPI, infoLogging)
}

func authenticateInfoLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req := c.Request()
		slog.Info(
			"audit",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("real_ip", c.RealIP()),
			slog.String("user_agent", req.UserAgent()),
			slog.String("uuid", c.Param("uuid")),
			slog.String("key", c.Param("key")),
		)

		return next(c)
	}
}

func infoLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		claims := authctx.Claims(c)
		req := c.Request()
		slog.Info(
			"audit",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("real_ip", c.RealIP()),
			slog.String("user_agent", req.UserAgent()),
			slog.String("uuid", c.Param("uuid")),
			slog.String("key", c.Param("key")),
			logutils.LogAny("query", c.QueryParams()),
			logutils.LogAny("claims", claims),
		)
		return next(c)
	}
}
