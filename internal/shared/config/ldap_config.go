package config

import (
	"fmt"
	"os"
	"strings"
)

type LDAPSetting struct {
	URL            string
	BaseDN         string
	BindDN         string
	BinddnPassword string
	SearchAttr     string
	SearchFilter   string
}

func GetLDAPSetting() (*LDAPSetting, error) {
	required := map[string]string{
		"LDAP_URL":              os.Getenv("LDAP_URL"),
		"LDAP_BASE_DN":            os.Getenv("LDAP_BASE_DN"),
		"LDAP_BIND_DN":            os.Getenv("LDAP_BIND_DN"),
		"LDAP_BIND_DN_PASSWORD":   os.Getenv("LDAP_BIND_DN_PASSWORD"),
		"LDAP_SEARCH_ATTR":        os.Getenv("LDAP_SEARCH_ATTR"),
		"LDAP_SEARCH_FILTER":      os.Getenv("LDAP_SEARCH_FILTER"),
	}
	for k, v := range required {
		if v == "" {
			return nil, fmt.Errorf("LDAP認証設定エラー: %s を環境変数に登録してください", k)
		}
	}
	return &LDAPSetting{
		URL:            required["LDAP_URL"],
		BaseDN:         required["LDAP_BASE_DN"],
		BindDN:         required["LDAP_BIND_DN"],
		BinddnPassword: required["LDAP_BIND_DN_PASSWORD"],
		SearchAttr:     required["LDAP_SEARCH_ATTR"],
		SearchFilter:   required["LDAP_SEARCH_FILTER"],
	}, nil
}

func GetSupporterDepartments() []string {
	raw := os.Getenv("SUPPORTER_DEPARTMENTS")
	if raw == "" {
		return []string{"総務", "人事"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func GetSystemUserEmail() string {
	if v := os.Getenv("SYSTEM_USER_EMAIL"); v != "" {
		return v
	}
	return "system@solvi.local"
}
