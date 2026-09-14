package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"solvi/internal/adapter/handler"
	ssehub "solvi/internal/adapter/handler/sse"
	"solvi/internal/application/converter"
	authuc "solvi/internal/application/usecase/auth_usecase"
	mnguc "solvi/internal/application/usecase/management_usecase"
	quc "solvi/internal/application/usecase/question_usecase"
	setuc "solvi/internal/application/usecase/setting_usecase"
	taguc "solvi/internal/application/usecase/tag_usecase"
	"solvi/internal/infrastructure/external/bedrock"
	"solvi/internal/infrastructure/external/ldap"
	"solvi/internal/infrastructure/postgresql"
	"solvi/internal/infrastructure/repository"
	"solvi/internal/shared/config"

	"github.com/joho/godotenv"
)

var (
	logDir                         = "logs"
	logFile, accessLogFile, awsLog *os.File
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	if _, err := os.Stat(logDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(logDir, 0755); err != nil {
				panic(err)
			}
		}
	}

	var err error
	logFile, err = os.OpenFile(time.Now().Format(logDir+`/log_20060102.log`), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}

	level := slog.LevelInfo
	mode := os.Getenv("MODE")
	if len(mode) >= 3 && strings.ToLower(mode[:3]) == "dev" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	accessLogFile, err = os.OpenFile(time.Now().Format(logDir+`/log_20060102_access.log`), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("failed to load timezone", "err", err)
		os.Exit(1)
	}
	time.Local = loc
	maxProc := runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	slog.Info("Use processor limit", "maxProc", maxProc)
}

const DATA_ROOT = "uploads"

func main() {
	defer func() {
		if err := accessLogFile.Close(); err != nil {
			slog.Error("failed to close access log file", "err", err)
			os.Exit(1)
		}
		if err := logFile.Close(); err != nil {
			panic(err.Error())
		}
	}()

	if err := os.MkdirAll(DATA_ROOT, 0755); err != nil {
		slog.Error("failed to create uploads dir", "err", err)
		os.Exit(1)
	}

	multiWriter := io.MultiWriter(os.Stdout, accessLogFile)
	ec := handler.NewEcho(multiWriter)

	db, err := postgresql.NewPostgresqlDB(config.GetPostgresqlDSN())
	if err != nil {
		slog.Error("failed to connect to postgresql", "err", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	tagRepo := repository.NewTagRepository(db)

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
	hub.RunSSE()

	questionUC := quc.NewQuestionUsecase(questionRepo, userRepo, bedrockClient, func(questionUUID string) {
		q, err := questionRepo.GetByUUID(questionUUID)
		if err != nil {
			return
		}
		detail := converter.QuestionEntityToDetail(q, true)
		hub.SendToQuestion("create-answer", detail, q.QuestionUserID)
		hub.SendToQuestion("update-question", detail, q.QuestionUserID)
	})

	deps := handler.Deps{
		Auth:       authuc.NewAuthUsecase(ldapClient, userRepo),
		Question:   questionUC,
		Setting:    setuc.NewSettingUsecase(userRepo, DATA_ROOT),
		Management: mnguc.NewManagementUsecase(userRepo),
		Tag:        taguc.NewTagUsecase(tagRepo),
		Hub:        hub,
	}
	handler.RegisterRoutes(ec, deps, multiWriter)

	if err := ec.Start(":" + config.GetEchoPort()); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
