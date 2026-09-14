package external

import (
	ldap_entity "solvi/internal/domain/entity/ldap"
)

//go:generate go tool mockgen -typed -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

type LDAPClient interface {
	Authenticate(userName, password string) (*ldap_entity.LDAPUser, error)
}
