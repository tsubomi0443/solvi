package lineworks

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	ext "solvi/internal/domain/interface/external"
	"solvi/internal/domain/valueobject"
	"solvi/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

const tokenRefreshSkew = 60 * time.Second

type Client struct {
	httpClient *http.Client
	cfg        config.LineWorksSetting
	key        *rsa.PrivateKey

	mu     sync.Mutex
	token  string
	expiry time.Time
	group  singleflight.Group
}

func NewClient(cfg config.LineWorksSetting) (*Client, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("LINE WORKSの秘密鍵を読み取れません")
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cfg:        cfg,
		key:        key,
	}, nil
}

func (c *Client) SendChannelMessage(ctx context.Context, text string) (int, error) {
	path := fmt.Sprintf("%s/v1.0/bots/%s/channels/%s/messages", c.cfg.APIBase, url.PathEscape(c.cfg.BotID), url.PathEscape(c.cfg.ChannelID))
	return c.send(ctx, path, text)
}

func (c *Client) SendUserMessage(ctx context.Context, userID, text string) (int, error) {
	path := fmt.Sprintf("%s/v1.0/bots/%s/users/%s/messages", c.cfg.APIBase, url.PathEscape(c.cfg.BotID), url.PathEscape(userID))
	return c.send(ctx, path, text)
}

func (c *Client) send(ctx context.Context, endpoint, text string) (int, error) {
	status, err := c.postMessage(ctx, endpoint, text, true)
	return status, err
}

func (c *Client) postMessage(ctx context.Context, endpoint, text string, allowRefresh bool) (int, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return 0, &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorTransient}
	}
	payload, err := json.Marshal(map[string]any{
		"content": map[string]string{
			"type": "text",
			"text": text,
		},
	})
	if err != nil {
		return 0, &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorTransient}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return 0, &ext.LineWorksSendError{Kind: valueobject.LineWorksErrorTransient}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		kind := valueobject.LineWorksErrorTransient
		if isTimeout(err) {
			kind = valueobject.LineWorksErrorUnknownOutcome
		}
		return 0, &ext.LineWorksSendError{Kind: kind}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized && allowRefresh {
		c.clearToken()
		return c.postMessage(ctx, endpoint, text, false)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp.StatusCode, nil
	}
	return resp.StatusCode, &ext.LineWorksSendError{
		Kind:       classifyStatus(resp.StatusCode),
		HTTPStatus: resp.StatusCode,
		RetryAfter: parseRateLimitReset(resp.Header.Get("RateLimit-Reset"), time.Now()),
	}
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Now().Before(c.expiry) {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	v, err, _ := c.group.Do("token", func() (any, error) {
		c.mu.Lock()
		if c.token != "" && time.Now().Before(c.expiry) {
			token := c.token
			c.mu.Unlock()
			return token, nil
		}
		c.mu.Unlock()
		token, expiry, err := c.fetchToken(ctx)
		if err != nil {
			return "", err
		}
		c.mu.Lock()
		c.token = token
		c.expiry = expiry
		c.mu.Unlock()
		return token, nil
	})
	if err != nil {
		return "", err
	}
	token, _ := v.(string)
	return token, nil
}

func (c *Client) clearToken() {
	c.mu.Lock()
	c.token = ""
	c.expiry = time.Time{}
	c.mu.Unlock()
}

func (c *Client) fetchToken(ctx context.Context) (string, time.Time, error) {
	assertion, err := c.signAssertion()
	if err != nil {
		return "", time.Time{}, err
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("scope", c.cfg.Scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", time.Time{}, fmt.Errorf("line works token request failed: status=%d", resp.StatusCode)
	}
	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("line works token request failed: status=%d", resp.StatusCode)
	}
	if parsed.ExpiresIn <= 0 {
		parsed.ExpiresIn = 3600
	}
	skew := tokenRefreshSkew
	if time.Duration(parsed.ExpiresIn)*time.Second <= skew*2 {
		skew = time.Duration(parsed.ExpiresIn) * time.Second / 2
	}
	expiry := time.Now().Add(time.Duration(parsed.ExpiresIn)*time.Second - skew)
	return parsed.AccessToken, expiry, nil
}

func (c *Client) signAssertion() (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": c.cfg.ClientID,
		"sub": c.cfg.ServiceAccount,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(c.key)
}

func classifyStatus(status int) valueobject.LineWorksErrorKind {
	switch {
	case status == http.StatusTooManyRequests:
		return valueobject.LineWorksErrorRateLimited
	case status == http.StatusForbidden:
		return valueobject.LineWorksErrorBotUnavailable
	case status >= 400 && status < 500:
		return valueobject.LineWorksErrorInvalidDestination
	default:
		return valueobject.LineWorksErrorTransient
	}
}

func parseRateLimitReset(raw string, now time.Time) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return 0
	}
	if v > 1_000_000_000 {
		d := time.Unix(v, 0).Sub(now)
		if d < 0 {
			return time.Second
		}
		return d
	}
	if v == 0 {
		return time.Second
	}
	return time.Duration(v) * time.Second
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}
