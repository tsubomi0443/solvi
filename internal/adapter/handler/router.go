package handler

import (
	"io"

	"solvi/internal/adapter/handler/api"
	"solvi/internal/adapter/handler/page"
	"solvi/internal/adapter/handler/sse"
	authuc "solvi/internal/application/usecase/auth_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"

	"github.com/labstack/echo/v5"
)

type Deps struct {
	Auth       *authuc.AuthUsecase
	Question   *quc.QuestionUsecase
	Setting    *setuc.SettingUsecase
	Management *mnguc.ManagementUsecase
	Tag        *taguc.TagUsecase
	Hub        *sse.Hub
}

func RegisterRoutes(e *echo.Echo, deps Deps, accessLog io.Writer) {
	e.Static("/static", "static")
	e.Static("/uploads", "uploads")

	ph := page.New(page.Deps{
		Question:   deps.Question,
		Setting:    deps.Setting,
		Management: deps.Management,
		Tag:        deps.Tag,
	})
	ah := api.New(api.Deps{
		Auth:       deps.Auth,
		Question:   deps.Question,
		Setting:    deps.Setting,
		Management: deps.Management,
		Tag:        deps.Tag,
		Hub:        deps.Hub,
	})
	sh := sse.NewHandler(deps.Hub)

	e.GET("/login", ph.LoginPage)
	e.POST("/api/v1/login", ah.Login)
	e.GET("/logout", ph.LogoutPage)

	auth := e.Group("", JWTConfig())

	auth.GET("/", ph.HomePage)
	auth.GET("/questions/new", ph.QuestionNewPage)
	auth.GET("/questions/:uuid", ph.QuestionDetailPage)
	auth.GET("/setting", ph.SettingPage)
	auth.GET("/tags", ph.TagsPage, SupporterOnlyPage)
	auth.GET("/management/users", ph.ManagementUsersPage, AdminOnly)
	auth.GET("/sse", sh.Stream)

	apiAuth := e.Group("/api/v1", APIJWTConfig())
	apiAuth.POST("/logout", ah.Logout)
	apiAuth.GET("/user/icon/:uuid", ah.GetUserIcon)
	apiAuth.GET("/questions", ah.ListQuestions)
	apiAuth.GET("/questions/:uuid", ah.GetQuestion)
	apiAuth.POST("/questions", ah.CreateQuestion)
	apiAuth.POST("/questions/:uuid/contents", ah.AppendContent)
	apiAuth.POST("/questions/:uuid/answers", ah.AddAnswer, SupporterOnly)
	apiAuth.DELETE("/questions/:uuid", ah.DeleteQuestion, AdminOnlyAPI)
	apiAuth.DELETE("/questions/:uuid/answers/:answerUuid", ah.DeleteAnswer, SupporterOrAdmin)
	apiAuth.POST("/questions/:uuid/memos", ah.AddMemo, SupporterOnly)
	apiAuth.DELETE("/questions/:uuid/memos/:memoUuid", ah.DeleteMemo, SupporterOrAdmin)
	apiAuth.POST("/questions/:uuid/refers", ah.AddRefer, SupporterOnly)
	apiAuth.DELETE("/questions/:uuid/refers/:referUuid", ah.DeleteRefer, SupporterOrAdmin)
	apiAuth.PUT("/questions/:uuid", ah.UpdateQuestion)
	apiAuth.PUT("/setting", ah.UpdateSetting)
	apiAuth.POST("/setting/icon", ah.UploadIcon)
	apiAuth.DELETE("/setting/icon", ah.DeleteIcon)
	apiAuth.GET("/management/users", ah.ListUsers, AdminOnlyAPI)
	apiAuth.PUT("/management/users/:uuid", ah.UpdateUser, AdminOnlyAPI)
	apiAuth.GET("/tags", ah.ListTags, SupporterOnly)
	apiAuth.PUT("/tags", ah.RenameTag, SupporterOnly)
	apiAuth.DELETE("/tags", ah.DeleteTag, SupporterOnly)
}
