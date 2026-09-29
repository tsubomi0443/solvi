package main

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"solvi/internal/adapter/handler"
	ssehub "solvi/internal/adapter/handler/sse"
	"solvi/internal/application/converter"
	authuc "solvi/internal/application/usecase/auth_usecase"
	lwuc "solvi/internal/application/usecase/lineworks_usecase"
	loguc "solvi/internal/application/usecase/log_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"
	"solvi/internal/infrastructure/external/bedrock"
	"solvi/internal/infrastructure/external/ldap"
	lwclient "solvi/internal/infrastructure/external/lineworks"
	applogger "solvi/internal/infrastructure/logger"
	"solvi/internal/infrastructure/postgresql"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/config"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	time.Local = loc
}

func main() {
	logDir := os.Getenv("LOG")
	if logDir == "" {
		logDir = "logs"
	}

	level := slog.LevelInfo
	if config.IsDevelop() {
		level = slog.LevelDebug
	}

	logs, err := applogger.New(applogger.Options{
		Dir:          logDir,
		Level:        level,
		AccessStdout: true,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logs.Close(); err != nil {
			slog.Error("failed to close log files", "err", err)
		}
	}()
	slog.SetDefault(logs.App)

	maxProc := runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	slog.Info("Use processor limit", "maxProc", maxProc)

	if err := os.MkdirAll(config.GetUploadDir(), 0755); err != nil {
		slog.Error("failed to create uploads dir", "err", err)
		os.Exit(1)
	}

	ec := handler.NewEcho(logs.Access, logs.Panic)

	db, err := postgresql.NewPostgresqlDB(config.GetPostgresqlDSN())
	if err != nil {
		slog.Error("failed to connect to postgresql", "err", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	tagRepo := repository.NewTagRepository(db)
	fileRepo := repository.NewFileRepository(config.GetUploadDir())
	logRepo, err := repository.NewLogRepository(logDir)
	if err != nil {
		slog.Error("failed to open log directory", "err", err)
		os.Exit(1)
	}

	ldapSetting, err := config.GetLDAPSetting()
	if err != nil {
		slog.Warn("LDAP not configured", "err", err)
	}
	var ldapClient *ldap.LDAPClient
	if ldapSetting != nil {
		ldapClient = ldap.NewClient(ldapSetting)
	}

	bedrockClient, err := bedrock.NewBedrockClient(context.Background())
	if err != nil {
		slog.Error("bedrock init failed", "err", err)
		os.Exit(1)
	}

	hub := ssehub.NewHub()
	go hub.Run()

	questionUC := quc.NewQuestionUsecase(questionRepo, userRepo, bedrockClient, func(questionUUID string) {
		q, err := questionRepo.GetByUUID(context.Background(), questionUUID)
		if err != nil {
			return
		}
		detail := converter.QuestionEntityToDetail(q, true)
		hub.SendToQuestion("create-answer", detail, q.QuestionUserID)
		hub.SendToQuestion("update-question", detail, q.QuestionUserID)
	})

	if lwSetting, ok := config.LoadLineWorks(); ok {
		lwClient, err := lwclient.NewClient(lwSetting)
		if err != nil {
			slog.Error("LINE WORKSの初期化に失敗しました", "err", err)
		} else {
			questionUC.ConfigureLineWorks(quc.LineWorksOptions{
				Enabled:    true,
				AppBaseURL: lwSetting.AppBaseURL,
				Debounce:   lwSetting.CommentDebounce,
			})
			worker := lwuc.New(repository.NewLineWorksNotificationRepository(db), lwClient)
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					if err := worker.ProcessDue(context.Background()); err != nil {
						slog.Error("LINE WORKS通知の送信に失敗しました", "err", err)
					}
				}
			}()
		}
	} else {
		slog.Warn("LINE WORKSは未設定です")
	}

	deps := handler.Deps{
		Auth:       authuc.NewAuthUsecase(ldapClient, userRepo),
		Question:   questionUC,
		Setting:    setuc.NewSettingUsecase(userRepo, fileRepo, config.GetUploadDir()),
		Management: mnguc.NewManagementUsecase(userRepo),
		Tag:        taguc.NewTagUsecase(tagRepo),
		Log:        loguc.NewLogUsecase(logRepo),
		Hub:        hub,
	}
	handler.RegisterRoutes(ec, deps, logs.Audit)

	if err := ec.Start(":" + config.GetEchoPort()); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
