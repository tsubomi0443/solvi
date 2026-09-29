package config

import (
	"os"
	"reflect"
	"testing"
)

func TestParseChannelIDs(t *testing.T) {
	tests := []struct {
		raw  string
		want []string
	}{
		{raw: "", want: nil},
		{raw: "   ", want: nil},
		{raw: "a", want: []string{"a"}},
		{raw: "a, b", want: []string{"a", "b"}},
		{raw: "a, b,,a", want: []string{"a", "b"}},
		{raw: " room1 , room2 ,room1", want: []string{"room1", "room2"}},
	}
	for _, tt := range tests {
		got := parseChannelIDs(tt.raw)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parseChannelIDs(%q)=%v want=%v", tt.raw, got, tt.want)
		}
	}
}

func TestLoadLineWorksWithoutChannelIDs(t *testing.T) {
	t.Setenv(LINEWORKS_CLIENT_ID, "cid")
	t.Setenv(LINEWORKS_CLIENT_SECRET, "secret")
	t.Setenv(LINEWORKS_SERVICE_ACCOUNT, "sa@cap")
	t.Setenv(LINEWORKS_PRIVATE_KEY, "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----")
	t.Setenv(LINEWORKS_BOT_ID, "bot")
	t.Setenv(LINEWORKS_CHANNEL_ID, "")
	t.Setenv(APP_BASE_URL, "https://solvi.example")

	cfg, ok := LoadLineWorks()
	if !ok {
		t.Fatal("expected LoadLineWorks to succeed without channel ids")
	}
	if len(cfg.ChannelIDs) != 0 {
		t.Fatalf("channel ids=%v", cfg.ChannelIDs)
	}
}

func TestLoadLineWorksParsesChannelCSV(t *testing.T) {
	t.Setenv(LINEWORKS_CLIENT_ID, "cid")
	t.Setenv(LINEWORKS_CLIENT_SECRET, "secret")
	t.Setenv(LINEWORKS_SERVICE_ACCOUNT, "sa@cap")
	t.Setenv(LINEWORKS_PRIVATE_KEY, "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----")
	t.Setenv(LINEWORKS_BOT_ID, "bot")
	t.Setenv(LINEWORKS_CHANNEL_ID, "room1, room2")
	t.Setenv(APP_BASE_URL, "https://solvi.example")

	cfg, ok := LoadLineWorks()
	if !ok {
		t.Fatal("expected LoadLineWorks to succeed")
	}
	want := []string{"room1", "room2"}
	if !reflect.DeepEqual(cfg.ChannelIDs, want) {
		t.Fatalf("channel ids=%v want=%v", cfg.ChannelIDs, want)
	}
}

func TestLoadLineWorksStillRequiresCoreEnv(t *testing.T) {
	for _, key := range []string{
		LINEWORKS_CLIENT_ID,
		LINEWORKS_CLIENT_SECRET,
		LINEWORKS_SERVICE_ACCOUNT,
		LINEWORKS_PRIVATE_KEY,
		LINEWORKS_BOT_ID,
		APP_BASE_URL,
	} {
		t.Run(key, func(t *testing.T) {
			os.Clearenv()
			t.Setenv(LINEWORKS_CLIENT_ID, "cid")
			t.Setenv(LINEWORKS_CLIENT_SECRET, "secret")
			t.Setenv(LINEWORKS_SERVICE_ACCOUNT, "sa@cap")
			t.Setenv(LINEWORKS_PRIVATE_KEY, "-----BEGIN RSA PRIVATE KEY-----\nkey\n-----END RSA PRIVATE KEY-----")
			t.Setenv(LINEWORKS_BOT_ID, "bot")
			t.Setenv(APP_BASE_URL, "https://solvi.example")
			t.Setenv(key, "")

			if _, ok := LoadLineWorks(); ok {
				t.Fatalf("expected LoadLineWorks to fail when %s is empty", key)
			}
		})
	}
}
