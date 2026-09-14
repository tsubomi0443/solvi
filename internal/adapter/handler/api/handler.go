package api

import (
	authuc "solvi/internal/application/usecase/auth_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"
	"solvi/internal/adapter/handler/sse"
)

type Deps struct {
	Auth       *authuc.AuthUsecase
	Question   *quc.QuestionUsecase
	Setting    *setuc.SettingUsecase
	Management *mnguc.ManagementUsecase
	Tag        *taguc.TagUsecase
	Hub        *sse.Hub
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}
