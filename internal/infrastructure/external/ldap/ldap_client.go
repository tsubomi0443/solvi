package ldap

import (
	"fmt"
	ldap_entity "solvi/internal/domain/entity/ldap"
	"solvi/internal/shared/config"

	"github.com/go-ldap/ldap/v3"
)

type LDAPClient struct {
	setting *config.LDAPSetting
}

func NewClient(setting *config.LDAPSetting) *LDAPClient {
	return &LDAPClient{setting: setting}
}

func (client *LDAPClient) newConnection() (*ldap.Conn, error) {
	return ldap.DialURL(client.setting.URL)
}

func (client *LDAPClient) bind(conn *ldap.Conn, bindDN, password string) error {
	if err := conn.Bind(bindDN, password); err != nil {
		return fmt.Errorf("コネクションのバインドに失敗しました: %w", err)
	}
	return nil
}

func (client *LDAPClient) Search(userName string) (*ldap_entity.LDAPUser, error) {
	conn, err := client.newConnection()
	if err != nil {
		return nil, fmt.Errorf("検索用接続に失敗しました: %w", err)
	}
	defer conn.Close()

	if err := client.bind(conn, client.setting.BindDN, client.setting.BinddnPassword); err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("(&%s(%s=%s))", client.setting.SearchFilter, client.setting.SearchAttr, ldap.EscapeFilter(userName))
	req := ldap.NewSearchRequest(client.setting.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, filter,
		[]string{"sAMAccountName", "displayName", "department", "mail", "distinguishedName", "employeeID"}, nil)
	result, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("LDAP検索に失敗しました: %w", err)
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("LDAP検索: 検索結果が0件でした")
	}
	e := result.Entries[0]
	return &ldap_entity.LDAPUser{
		UserName:   e.GetAttributeValue("sAMAccountName"),
		Name:       e.GetAttributeValue("displayName"),
		Department: e.GetAttributeValue("department"),
		EmployeeID: e.GetAttributeValue("employeeID"),
		Mail:       e.GetAttributeValue("mail"),
		DN:         e.GetAttributeValue("distinguishedName"),
	}, nil
}

func (client *LDAPClient) Authenticate(userName, password string) (*ldap_entity.LDAPUser, error) {
	user, err := client.Search(userName)
	if err != nil {
		return nil, err
	}
	conn, err := client.newConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := client.bind(conn, user.DN, password); err != nil {
		return nil, fmt.Errorf("パスワード認証に失敗: %w", err)
	}
	return user, nil
}
