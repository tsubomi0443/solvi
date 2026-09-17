package api

import (
	"log/slog"
	"solvi/internal/adapter/handler/sse"
	authuc "solvi/internal/application/usecase/auth_usecase"
	loguc "solvi/internal/application/usecase/log_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"
)

type Deps struct {
	Auth       *authuc.AuthUsecase
	Question   *quc.QuestionUsecase
	Setting    *setuc.SettingUsecase
	Management *mnguc.ManagementUsecase
	Tag        *taguc.TagUsecase
	Log        *loguc.LogUsecase
	Hub        *sse.Hub
	Audit      *slog.Logger
}

type Handler struct {
	deps Deps
}

func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}
