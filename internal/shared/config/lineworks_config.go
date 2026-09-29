package config

import (
	"os"
	"strconv"
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
	NOTICE_DATE                = "NOTICE_DATE"
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
	ChannelIDs      []string
	APIBase         string
	TokenURL        string
	Scope           string
	AppBaseURL      string
	CommentDebounce time.Duration
	NoticeTimes     []NoticeTime
}

// NoticeTime は NOTICE_DATE の hh:mm（ゼロ埋め2桁）を表す。
type NoticeTime struct {
	Hour   int
	Minute int
}

func (t NoticeTime) String() string {
	return formatNoticeTime(t.Hour, t.Minute)
}

func (t NoticeTime) Matches(now time.Time, loc *time.Location) bool {
	in := now.In(loc)
	return in.Hour() == t.Hour && in.Minute() == t.Minute
}

func (t NoticeTime) Burst() int {
	return t.Hour*60 + t.Minute
}

// LoadLineWorks は通知に必要な環境変数が揃っているときだけ設定を返す。
// RedirectURL は保持するが、Service Account の JWT 取得では使わない。
func LoadLineWorks() (LineWorksSetting, bool) {
	clientID := strings.TrimSpace(os.Getenv(LINEWORKS_CLIENT_ID))
	clientSecret := os.Getenv(LINEWORKS_CLIENT_SECRET)
	serviceAccount := strings.TrimSpace(os.Getenv(LINEWORKS_SERVICE_ACCOUNT))
	privateKey := normalizePEM(os.Getenv(LINEWORKS_PRIVATE_KEY))
	botID := strings.TrimSpace(os.Getenv(LINEWORKS_BOT_ID))
	channelIDs := parseChannelIDs(os.Getenv(LINEWORKS_CHANNEL_ID))
	appBase := strings.TrimRight(strings.TrimSpace(os.Getenv(APP_BASE_URL)), "/")
	if clientID == "" || strings.TrimSpace(clientSecret) == "" || serviceAccount == "" || privateKey == "" || botID == "" || appBase == "" {
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
		ChannelIDs:      channelIDs,
		APIBase:         apiBase,
		TokenURL:        tokenURL,
		Scope:           scope,
		AppBaseURL:      appBase,
		CommentDebounce: commentDebounce(),
		NoticeTimes:     parseNoticeTimes(os.Getenv(NOTICE_DATE)),
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

// parseChannelIDs は LINEWORKS_CHANNEL_ID の CSV を trim し、空要素と重複を除いて返す。
func parseChannelIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseNoticeTimes は NOTICE_DATE の CSV を hh:mm（ゼロ埋め2桁）として解釈する。
func parseNoticeTimes(raw string) []NoticeTime {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := make(map[int]struct{})
	out := make([]NoticeTime, 0)
	for _, part := range strings.Split(raw, ",") {
		t, ok := parseNoticeTime(strings.TrimSpace(part))
		if !ok {
			continue
		}
		key := t.Burst()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseNoticeTime(raw string) (NoticeTime, bool) {
	if len(raw) != 5 || raw[2] != ':' {
		return NoticeTime{}, false
	}
	h, err := strconv.Atoi(raw[:2])
	if err != nil || h < 0 || h > 23 {
		return NoticeTime{}, false
	}
	m, err := strconv.Atoi(raw[3:])
	if err != nil || m < 0 || m > 59 {
		return NoticeTime{}, false
	}
	return NoticeTime{Hour: h, Minute: m}, true
}

func formatNoticeTime(hour, minute int) string {
	return pad2(hour) + ":" + pad2(minute)
}

func pad2(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}
