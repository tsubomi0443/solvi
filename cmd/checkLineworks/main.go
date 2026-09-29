// checkLineworks は LINE WORKS 通知設定の診断 CLI です。
// 引数なし: 環境変数と Bot API で設定済みトークルームを照合します。
// listen: Callback イベントを待ち受け、channelId / userId などを表示します。
package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"solvi/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

const defaultListenAddr = ":18080"

func main() {
	_ = godotenv.Load(".env")

	if len(os.Args) > 1 && os.Args[1] == "listen" {
		if err := runListen(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 {
		printUsage()
		os.Exit(1)
	}

	if err := runCheck(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/checkLineworks          # 設定と Bot API を照合")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/checkLineworks listen   # Callback 待ち受け (既定 :18080/callback)")
}

func runListen() error {
	cfg, ok := config.LoadLineWorks()
	if !ok {
		return fmt.Errorf("LINE WORKS 設定が不足しています: %s", missingLineWorksEnv())
	}

	addr := defaultListenAddr
	if v := strings.TrimSpace(os.Getenv("LINEWORKS_CHECK_LISTEN_ADDR")); v != "" {
		addr = v
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if botID := strings.TrimSpace(r.Header.Get("X-WORKS-BotId")); botID != "" && botID != cfg.BotID {
			fmt.Fprintf(os.Stderr, "warning: X-WORKS-BotId=%s (expected %s)\n", botID, cfg.BotID)
		}

		var evt callbackEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid JSON: %v\n", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Printf("[%s] type=%s", time.Now().Format(time.RFC3339), evt.Type)
		if evt.Source.UserID != "" {
			fmt.Printf(" userId=%s", evt.Source.UserID)
		}
		if evt.Source.ChannelID != "" {
			fmt.Printf(" channelId=%s", evt.Source.ChannelID)
			if channelConfigured(cfg, evt.Source.ChannelID) {
				fmt.Print(" (matches LINEWORKS_CHANNEL_ID)")
			}
		}
		if evt.Source.DomainID != 0 {
			fmt.Printf(" domainId=%d", evt.Source.DomainID)
		}
		if len(evt.Members) > 0 {
			fmt.Printf(" members=%d [%s]", len(evt.Members), strings.Join(evt.Members, ", "))
		}
		fmt.Println()

		if evt.Source.ChannelID != "" && !channelConfigured(cfg, evt.Source.ChannelID) {
			fmt.Printf("  hint: add to LINEWORKS_CHANNEL_ID=%s\n", evt.Source.ChannelID)
		}

		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	fmt.Printf("listening on http://127.0.0.1%s/callback (Bot ID=%s)\n", addr, cfg.BotID)
	fmt.Println("Developer Console の Callback URL をこの URL に向け、Bot をトークルームへ招待するかメッセージを送ってください。")
	fmt.Println("Ctrl+C で終了します。")

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

func runCheck() error {
	cfg, ok := config.LoadLineWorks()
	if !ok {
		return fmt.Errorf("LINE WORKS 設定が不足しています: %s", missingLineWorksEnv())
	}

	printConfigSummary(cfg)

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	if err != nil {
		return fmt.Errorf("秘密鍵を読み取れません: %w", err)
	}
	fmt.Println("ok: 秘密鍵 (RSA)")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, expiresIn, err := fetchAccessToken(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("アクセストークン取得失敗: %w", err)
	}
	_ = token
	fmt.Printf("ok: アクセストークン取得 (expires_in=%ds)\n", expiresIn)

	client := &apiClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cfg:        cfg,
		token:      token,
	}

	if len(cfg.ChannelIDs) == 0 {
		fmt.Println("skip: LINEWORKS_CHANNEL_ID 未設定のためトークルーム確認を省略")
	} else {
		for i, chID := range cfg.ChannelIDs {
			if i > 0 {
				fmt.Println()
			}
			if err := checkChannel(ctx, client, cfg, chID); err != nil {
				return err
			}
		}
	}

	bot, status, err := client.getBot(ctx, cfg.BotID)
	if err != nil {
		if status == http.StatusForbidden {
			fmt.Printf("warning: Bot 情報取得は HTTP 403 (scope=%s)。bot / bot.read が必要な場合があります\n", cfg.Scope)
		} else {
			return fmt.Errorf("Bot 情報取得失敗 (HTTP %d): %w", status, err)
		}
	} else {
		fmt.Println("--- Bot ---")
		fmt.Printf("botId=%s\n", bot.BotIDString())
		fmt.Printf("botName=%s\n", bot.BotName)
		if bot.BotIDString() == cfg.BotID {
			fmt.Println("ok: botId が LINEWORKS_BOT_ID と一致")
		} else {
			fmt.Printf("warning: 応答 botId=%s が LINEWORKS_BOT_ID=%s と不一致\n", bot.BotIDString(), cfg.BotID)
		}
	}

	fmt.Println("診断完了")
	return nil
}

func missingLineWorksEnv() string {
	required := []struct {
		name string
		ok   func() bool
	}{
		{config.LINEWORKS_CLIENT_ID, func() bool { return strings.TrimSpace(os.Getenv(config.LINEWORKS_CLIENT_ID)) != "" }},
		{config.LINEWORKS_CLIENT_SECRET, func() bool { return strings.TrimSpace(os.Getenv(config.LINEWORKS_CLIENT_SECRET)) != "" }},
		{config.LINEWORKS_SERVICE_ACCOUNT, func() bool { return strings.TrimSpace(os.Getenv(config.LINEWORKS_SERVICE_ACCOUNT)) != "" }},
		{config.LINEWORKS_PRIVATE_KEY, func() bool { return strings.TrimSpace(os.Getenv(config.LINEWORKS_PRIVATE_KEY)) != "" }},
		{config.LINEWORKS_BOT_ID, func() bool { return strings.TrimSpace(os.Getenv(config.LINEWORKS_BOT_ID)) != "" }},
		{config.APP_BASE_URL, func() bool { return strings.TrimSpace(os.Getenv(config.APP_BASE_URL)) != "" }},
	}
	var missing []string
	for _, r := range required {
		if !r.ok() {
			missing = append(missing, r.name)
		}
	}
	return strings.Join(missing, ", ")
}

func channelConfigured(cfg config.LineWorksSetting, channelID string) bool {
	for _, id := range cfg.ChannelIDs {
		if id == channelID {
			return true
		}
	}
	return false
}

func checkChannel(ctx context.Context, client *apiClient, cfg config.LineWorksSetting, chID string) error {
	channel, status, err := client.getChannel(ctx, cfg.BotID, chID)
	if err != nil {
		return fmt.Errorf("トークルーム取得失敗 channelId=%s (HTTP %d): %w", chID, status, err)
	}
	fmt.Println("--- トークルーム ---")
	fmt.Printf("configuredChannelId=%s\n", chID)
	fmt.Printf("domainId=%d\n", channel.DomainID)
	fmt.Printf("channelId=%s\n", channel.ChannelID)
	fmt.Printf("title=%s\n", channel.Title)
	fmt.Printf("channelType=%s\n", channel.ChannelType.Type)
	if channel.ChannelType.OrgUnitID != "" {
		fmt.Printf("orgUnitId=%s\n", channel.ChannelType.OrgUnitID)
	}
	if channel.ChannelType.GroupID != "" {
		fmt.Printf("groupId=%s\n", channel.ChannelType.GroupID)
	}
	if channel.ChannelID == chID {
		fmt.Println("ok: channelId が LINEWORKS_CHANNEL_ID と一致")
	} else {
		fmt.Printf("warning: 応答 channelId=%s が LINEWORKS_CHANNEL_ID=%s と不一致\n", channel.ChannelID, chID)
	}

	members, err := client.listAllMembers(ctx, cfg.BotID, chID)
	if err != nil {
		return fmt.Errorf("メンバー一覧取得失敗 channelId=%s: %w", chID, err)
	}
	fmt.Println("--- メンバー ---")
	fmt.Printf("count=%d\n", len(members))
	for _, m := range members {
		fmt.Printf("  %s\n", m)
	}
	return nil
}

func printConfigSummary(cfg config.LineWorksSetting) {
	fmt.Println("--- 設定 ---")
	fmt.Printf("clientId=%s\n", cfg.ClientID)
	fmt.Printf("serviceAccount=%s\n", cfg.ServiceAccount)
	fmt.Printf("botId=%s\n", cfg.BotID)
	if len(cfg.ChannelIDs) == 0 {
		fmt.Println("channelIds=(none)")
	} else {
		fmt.Printf("channelIds=%s\n", strings.Join(cfg.ChannelIDs, ", "))
	}
	fmt.Printf("apiBase=%s\n", cfg.APIBase)
	fmt.Printf("tokenUrl=%s\n", cfg.TokenURL)
	fmt.Printf("scope=%s\n", cfg.Scope)
	fmt.Printf("appBaseUrl=%s\n", cfg.AppBaseURL)
}

type callbackEvent struct {
	Type    string `json:"type"`
	Source  source `json:"source"`
	Members []string `json:"members"`
}

type source struct {
	UserID    string `json:"userId"`
	ChannelID string `json:"channelId"`
	DomainID  int64  `json:"domainId"`
}

type apiClient struct {
	httpClient *http.Client
	cfg        config.LineWorksSetting
	token      string
}

type channelInfo struct {
	DomainID    int64       `json:"domainId"`
	ChannelID   string      `json:"channelId"`
	Title       string      `json:"title"`
	ChannelType channelType `json:"channelType"`
}

type channelType struct {
	Type      string `json:"type"`
	OrgUnitID string `json:"orgUnitId"`
	GroupID   string `json:"groupId"`
}

type botInfo struct {
	BotID   json.Number `json:"botId"`
	BotName string      `json:"botName"`
}

func (b botInfo) BotIDString() string {
	return b.BotID.String()
}

type membersResponse struct {
	Members          []string `json:"members"`
	ResponseMetaData struct {
		NextCursor string `json:"nextCursor"`
	} `json:"responseMetaData"`
}

func fetchAccessToken(ctx context.Context, cfg config.LineWorksSetting, key *rsa.PrivateKey) (string, int, error) {
	assertion, err := signAssertion(cfg, key)
	if err != nil {
		return "", 0, err
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("scope", cfg.Scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", 0, fmt.Errorf("status=%d", resp.StatusCode)
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.AccessToken == "" {
		return "", 0, fmt.Errorf("invalid token response")
	}
	if parsed.ExpiresIn <= 0 {
		parsed.ExpiresIn = 3600
	}
	return parsed.AccessToken, parsed.ExpiresIn, nil
}

func signAssertion(cfg config.LineWorksSetting, key *rsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": cfg.ClientID,
		"sub": cfg.ServiceAccount,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func (c *apiClient) get(ctx context.Context, path string, dest any) (int, error) {
	endpoint := fmt.Sprintf("%s/v1.0/%s", c.cfg.APIBase, strings.TrimPrefix(path, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}
	if dest == nil {
		return resp.StatusCode, nil
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

func (c *apiClient) getChannel(ctx context.Context, botID, channelID string) (channelInfo, int, error) {
	path := fmt.Sprintf("bots/%s/channels/%s", url.PathEscape(botID), url.PathEscape(channelID))
	var ch channelInfo
	status, err := c.get(ctx, path, &ch)
	return ch, status, err
}

func (c *apiClient) getBot(ctx context.Context, botID string) (botInfo, int, error) {
	path := fmt.Sprintf("bots/%s", url.PathEscape(botID))
	var bot botInfo
	status, err := c.get(ctx, path, &bot)
	return bot, status, err
}

func (c *apiClient) listAllMembers(ctx context.Context, botID, channelID string) ([]string, error) {
	var all []string
	cursor := ""
	for {
		path := fmt.Sprintf("bots/%s/channels/%s/members?count=100", url.PathEscape(botID), url.PathEscape(channelID))
		if cursor != "" {
			path += "&cursor=" + url.QueryEscape(cursor)
		}
		var resp membersResponse
		status, err := c.get(ctx, path, &resp)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d: %w", status, err)
		}
		all = append(all, resp.Members...)
		cursor = strings.TrimSpace(resp.ResponseMetaData.NextCursor)
		if cursor == "" {
			break
		}
	}
	return all, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
