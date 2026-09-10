package main

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

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
	if strings.ToLower(os.Getenv("MODE")[:2]) == "dev" {
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

const (
	DATA_ROOT = "uploads"
)

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

	// multiWriter := io.MultiWriter(os.Stdout, accessLogFile)
	// ec = handler.NewEcho(multiWriter)
	// dp, err := infrastructure.NewPostgresqlDB("")
	// if err != nil {
	// 	slog.Error("failed to connect to postgresql", "err", err)
	// 	os.Exit(1)
	// }

	os.Exit(0)
}
