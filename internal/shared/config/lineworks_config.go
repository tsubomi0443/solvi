package config

import (
	"os"
	"strings"
	"time"
)

const (
	LINEWORKS_CLIENT_ID        = "LINEWORKS_CLIENT_ID"
	LINEWORKS_CLIENT_SECRET    = "LINEWORKS_CLIENT_SECRET"
	LINEWORKS_REDIRECT_URL     = "LINEWORKS_REDIRECT_URL"
	LINEWORKS_SERVICE_ACCOUNT  = "LINEWORKS_SERVICE_ACCOUNT"
	LINEWORKS_PRIVATE_KEY      = "LINEWORKS_PRIVATE_KEY"
	LINEWORKS_BOT_ID           = "LINEWORKS_BOT_ID"
	LINEWORKS_CHANNEL_ID       = "LINEWORKS_CHANNEL_ID"
	LINEWORKS_API_BASE         = "LINEWORKS_API_BASE"
	LINEWORKS_TOKEN_URL        = "LINEWORKS_TOKEN_URL"
	LINEWORKS_SCOPE            = "LINEWORKS_SCOPE"
	LINEWORKS_COMMENT_DEBOUNCE = "LINEWORKS_COMMENT_DEBOUNCE"
	APP_BASE_URL               = "APP_BASE_URL"

	defaultLineWorksAPIBase  = "https://www.worksapis.com"
	defaultLineWorksTokenURL = "https://auth.worksmobile.com/oauth2/v2.0/token"
	defaultLineWorksScope    = "bot.message"
	defaultCommentDebounce   = 2 * time.Minute
)

// LineWorksSetting は LINE WORKS Bot API 2.0 の接続情報。秘密値はログに出さない。
type LineWorksSetting struct {
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	ServiceAccount  string
	PrivateKey      string
	BotID           string
	ChannelID       string
	APIBase         string
	TokenURL        string
	Scope           string
	AppBaseURL      string
	CommentDebounce time.Duration
}

// LoadLineWorks は通知に必要な環境変数が揃っているときだけ設定を返す。
// RedirectURL は保持するが、Service Account の JWT 取得では使わない。
func LoadLineWorks() (LineWorksSetting, bool) {
	clientID := strings.TrimSpace(os.Getenv(LINEWORKS_CLIENT_ID))
	clientSecret := os.Getenv(LINEWORKS_CLIENT_SECRET)
	serviceAccount := strings.TrimSpace(os.Getenv(LINEWORKS_SERVICE_ACCOUNT))
	privateKey := normalizePEM(os.Getenv(LINEWORKS_PRIVATE_KEY))
	botID := strings.TrimSpace(os.Getenv(LINEWORKS_BOT_ID))
	channelID := strings.TrimSpace(os.Getenv(LINEWORKS_CHANNEL_ID))
	appBase := strings.TrimRight(strings.TrimSpace(os.Getenv(APP_BASE_URL)), "/")
	if clientID == "" || strings.TrimSpace(clientSecret) == "" || serviceAccount == "" || privateKey == "" || botID == "" || channelID == "" || appBase == "" {
		return LineWorksSetting{}, false
	}

	apiBase := strings.TrimRight(strings.TrimSpace(os.Getenv(LINEWORKS_API_BASE)), "/")
	if apiBase == "" {
		apiBase = defaultLineWorksAPIBase
	}
	tokenURL := strings.TrimSpace(os.Getenv(LINEWORKS_TOKEN_URL))
	if tokenURL == "" {
		tokenURL = defaultLineWorksTokenURL
	}
	scope := strings.TrimSpace(os.Getenv(LINEWORKS_SCOPE))
	if scope == "" {
		scope = defaultLineWorksScope
	}

	return LineWorksSetting{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURL:     strings.TrimSpace(os.Getenv(LINEWORKS_REDIRECT_URL)),
		ServiceAccount:  serviceAccount,
		PrivateKey:      privateKey,
		BotID:           botID,
		ChannelID:       channelID,
		APIBase:         apiBase,
		TokenURL:        tokenURL,
		Scope:           scope,
		AppBaseURL:      appBase,
		CommentDebounce: commentDebounce(),
	}, true
}

func commentDebounce() time.Duration {
	raw := strings.TrimSpace(os.Getenv(LINEWORKS_COMMENT_DEBOUNCE))
	if raw == "" {
		return defaultCommentDebounce
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultCommentDebounce
	}
	return d
}

func normalizePEM(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"`)
	return strings.ReplaceAll(raw, `\n`, "\n")
}
