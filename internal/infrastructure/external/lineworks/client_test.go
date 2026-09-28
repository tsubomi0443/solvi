package lineworks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ext "solvi/internal/domain/interface/external"
	"solvi/internal/domain/valueobject"
	"solvi/internal/shared/config"
)

func testKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
}

func TestClientTokenRefreshIsSingleFlightAndSecretsStayOutOfErrors(t *testing.T) {
	var tokenCalls atomic.Int32
	const secret = "super-secret-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			if tokenCalls.Add(1) == 1 {
				time.Sleep(50 * time.Millisecond)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, secret, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client, err := NewClient(config.LineWorksSetting{
		ClientID:       "cid",
		ClientSecret:   secret,
		ServiceAccount: "sa",
		PrivateKey:     testKey(t),
		BotID:          "bot",
		ChannelID:      "room",
		APIBase:        srv.URL,
		TokenURL:       srv.URL + "/token",
		Scope:          "bot.message",
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.SendChannelMessage(context.Background(), "hello")
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("token calls=%d", tokenCalls.Load())
	}
}

func TestClientRateLimitAndPermanentErrors(t *testing.T) {
	const secret = "super-secret-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
			return
		}
		if strings.Contains(r.URL.Path, "/users/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("RateLimit-Reset", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client, err := NewClient(config.LineWorksSetting{
		ClientID: "cid", ClientSecret: secret, ServiceAccount: "sa", PrivateKey: testKey(t),
		BotID: "bot", ChannelID: "room", APIBase: srv.URL, TokenURL: srv.URL + "/token", Scope: "bot.message",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SendChannelMessage(context.Background(), "hello")
	sendErr, ok := err.(*ext.LineWorksSendError)
	if !ok || sendErr.Kind != valueobject.LineWorksErrorRateLimited || sendErr.RetryAfter != 30*time.Second {
		t.Fatalf("err=%v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal(err)
	}

	_, err = client.SendUserMessage(context.Background(), "tanaka", "hello")
	sendErr, ok = err.(*ext.LineWorksSendError)
	if !ok || sendErr.Kind != valueobject.LineWorksErrorInvalidDestination || sendErr.HTTPStatus != http.StatusNotFound {
		t.Fatalf("err=%v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal(err)
	}
}

func TestParseRateLimitResetUnix(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	got := parseRateLimitReset("1700000030", now)
	if got != 30*time.Second {
		t.Fatalf("got=%s", got)
	}
}
