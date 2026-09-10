package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	E_PORT   = "PORT"
	FONT     = "FONT"
	JWT_KEY  = "JWK_KEY"
	PEPPER   = "PEPPER"
	MODE_IS  = "MODE"
	HOST     = "PSQL_HOST"
	D_PORT   = "PSQL_PORT"
	USER     = "PSQL_USER"
	PASSWORD = "PSQL_PASSWORD"
	DB_NAME  = "PSQL_DB_NAME"
	SSL_MODE = "PSQL_SSLMODE"
	LOG_PATH = "LOG"

	COGNITO_DOMAIN = "COGNITO_DOMAIN"
	IDP            = "IDP"
	CLIENT_ID      = "CLIENT_ID"
	CLIENT_SECRET  = "CLIENT_SECRET"
	REDIRECT_URI   = "REDIRECT_URI"
	SCOPE          = "SCOPE"
	LA             = "LA"
	ISSUER_URL     = "ISSUER_URL"

	AWS_REDIS_ENDPOINT = "AWS_REDIS_ENDPOINT"
	AWS_REDIS_PASSWORD = "AWS_REDIS_PASSWORD"
	AWS_REDIS_DB_NUM   = "AWS_REDIS_DB_NUM"
	AWS_REDIS_POOLSIZE = "AWS_REDIS_POOLSIZE"

	AWS_REGION                   = "AWS_REGION"
	AWS_S3_ARN                   = "AWS_S3_ARN"
	AWS_S3_BUCKET                = "AWS_S3_BUCKET"
	AWS_S3_BUCKET_KEY            = "AWS_S3_BUCKET_KEY"
	AWS_BEDROCK_DATASOURCE_ID    = "AWS_BEDROCK_DATASOURCE_ID"
	AWS_BEDROCK_KNOWLEDGEBASE_ID = "AWS_BEDROCK_KNOWLEDGEBASE_ID"
	AWS_BEDROCK_AGENT_ID         = "AWS_BEDROCK_AGENT_ID"
	AWS_BEDROCK_AGENT_ALIAS_ID   = "AWS_BEDROCK_AGENT_ALIAS_ID"
	AWS_LL_MODEL_ARN             = "AWS_LL_MODEL_ARN"
)

// GetDSN は環境変数から PostgreSQL 接続文字列を組み立てる（.env.example 参照）。
//
// return:
//   - string: lib/pq 形式の DSN
func GetPostgresqlDSN() string {
	const dsnTemplate string = "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Tokyo"
	_host, err := getHost()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	_user, err := getUser()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	_password, err := getPassword()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	_dbname, err := getDBname()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	_port, err := getDatabasePort()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	_sslmode, err := getSSLMode()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	return fmt.Sprintf(dsnTemplate, _host, _user, _password, _dbname, _port, _sslmode)
}

func getHost() (string, error) {
	_host := os.Getenv(HOST)
	if _host == "" {
		return "", fmt.Errorf("configure to %s", HOST)
	}
	return _host, nil
}

func getDatabasePort() (string, error) {
	_port := os.Getenv(D_PORT)
	if _port == "" {
		return "", fmt.Errorf("configure to %s", D_PORT)
	}
	return _port, nil
}

func getUser() (string, error) {
	_user := os.Getenv(USER)
	if _user == "" {
		return "", fmt.Errorf("configure to %s", USER)
	}
	return _user, nil
}

func getPassword() (string, error) {
	_password := os.Getenv(PASSWORD)
	if _password == "" {
		return "", fmt.Errorf("configure to %s", PASSWORD)
	}
	return _password, nil
}

func getDBname() (string, error) {
	_dbname := os.Getenv(DB_NAME)
	if _dbname == "" {
		return "", fmt.Errorf("configure to %s", DB_NAME)
	}
	return _dbname, nil
}

func getSSLMode() (string, error) {
	_sslmode := os.Getenv(SSL_MODE)
	if _sslmode == "" {
		return "", fmt.Errorf("configure to %s", SSL_MODE)
	}
	return _sslmode, nil
}

func GetEchoPort() string {
	_port := os.Getenv(E_PORT)
	if _port == "" {
		slog.Error("configuration error", "key", E_PORT)
		os.Exit(1)
	}
	return _port
}

// GetJWTKey は JWT 署名用の秘密鍵を返す。未設定時はプロセスを終了する。
//
// return:
//   - string: JWT 署名鍵
func GetJWTKey() string {
	_jwtKey := os.Getenv(JWT_KEY)
	if _jwtKey == "" {
		slog.Error("configuration error", "key", JWT_KEY)
		os.Exit(1)
	}
	return _jwtKey
}

// IsDevelop は開発モード（MODE=DEV/DEVELOP）かどうかを返す。
//
// return:
//   - bool: 開発モードなら true
func IsDevelop() bool {
	switch strings.ToUpper(os.Getenv(MODE_IS)) {
	case "DEV", "DEVELOP":
		return true
	default:
		return false
	}
}

// IsProduct は本番モード（MODE=PRO/PROD/PRODUCT）かどうかを返す。
//
// return:
//   - bool: 本番モードなら true
func IsProduct() bool {
	switch strings.ToUpper(os.Getenv(MODE_IS)) {
	case "PRO", "PROD", "PRODUCT":
		return true
	default:
		return false
	}
}

// GetPepper はパスワードハッシュ（Argon2 / SHA 補助）用のアプリケーションペッパーを返す。未設定時はプロセスを終了する。
//
// return:
//   - string: ペッパー文字列
func GetPepper() string {
	_pepper := os.Getenv(PEPPER)
	if _pepper == "" {
		slog.Error("configuration error", "key", PEPPER)
		os.Exit(1)
	}
	return _pepper
}

func GetLogDir() string {
	_log := os.Getenv(LOG_PATH)
	if _log == "" {
		slog.Error("configuration error", "key", LOG_PATH)
		os.Exit(1)
	}
	return _log
}

func GetCognitoURI() string {
	uri := strings.Builder{}
	v, err := getCognitoDomain()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	v, err = getIdp()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	v, err = GetClientID()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	v, err = GetRedirectURI()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	v, err = getScope()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	v, err = getLa()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}
	uri.WriteString(v)

	return uri.String()
}

func getCognitoDomain() (string, error) {
	cognitoDomain := os.Getenv(COGNITO_DOMAIN)
	if cognitoDomain == "" {
		return "", fmt.Errorf("configure to %s", COGNITO_DOMAIN)
	}
	return cognitoDomain, nil
}

func getIdp() (string, error) {
	idp := os.Getenv(IDP)
	if idp == "" {
		return "", fmt.Errorf("configure to %s", IDP)
	}
	return idp, nil
}

func GetClientID() (string, error) {
	clientID := os.Getenv(CLIENT_ID)
	if clientID == "" {
		return "", fmt.Errorf("configure to %s", CLIENT_ID)
	}
	return clientID, nil
}

func GetClientSecret() (string, error) {
	clientSecret := os.Getenv(CLIENT_SECRET)
	if clientSecret == "" {
		return "", fmt.Errorf("configure to %s", CLIENT_SECRET)
	}
	return clientSecret, nil
}

func GetRedirectURI() (string, error) {
	redirectURI := os.Getenv(REDIRECT_URI)
	if redirectURI == "" {
		return "", fmt.Errorf("configure to %s", REDIRECT_URI)
	}
	return redirectURI, nil
}

func getScope() (string, error) {
	scope := os.Getenv(SCOPE)
	if scope == "" {
		return "", fmt.Errorf("configure to %s", SCOPE)
	}
	return scope, nil
}

func getLa() (string, error) {
	la := os.Getenv(LA)
	if la == "" {
		return "", fmt.Errorf("configure to %s", LA)
	}
	return la, nil
}

func GetIssuerURL() (string, error) {
	issuerURL := os.Getenv(ISSUER_URL)
	if issuerURL == "" {
		return "", fmt.Errorf("configure to %s", ISSUER_URL)
	}
	return issuerURL, nil
}

func GetRedisEndpoint() string {
	_endpoint := os.Getenv(AWS_REDIS_ENDPOINT)
	if _endpoint == "" {
		slog.Error("configuration error", "key", AWS_REDIS_ENDPOINT)
		os.Exit(1)
	}
	return _endpoint
}

func GetRedisPassword() string {
	_password := os.Getenv(AWS_REDIS_PASSWORD)
	if _password == "" {
		slog.Error("configuration error", "key", AWS_REDIS_PASSWORD)
		os.Exit(1)
	}
	return _password
}

func GetRedisDB() string {
	_dbNum := os.Getenv(AWS_REDIS_DB_NUM)
	if _dbNum == "" {
		slog.Error("configuration error", "key", AWS_REDIS_DB_NUM)
		os.Exit(1)
	}
	return _dbNum
}

func GetRedisPoolsize() string {
	_poolsize := os.Getenv(AWS_REDIS_POOLSIZE)
	if _poolsize == "" {
		slog.Error("configuration error", "key", AWS_REDIS_POOLSIZE)
		os.Exit(1)
	}
	return _poolsize
}

func GetFont() string {
	_font := os.Getenv(FONT)
	if _font == "" {
		slog.Error("configuration error", "key", FONT)
		os.Exit(1)
	}
	return _font
}

func GetRegion() string {
	_region := os.Getenv(AWS_REGION)
	if _region == "" {
		slog.Error("configuration error", "key", AWS_REGION)
		os.Exit(1)
	}
	return _region
}

func GetS3Arn() string {
	_arn := os.Getenv(AWS_S3_ARN)
	if _arn == "" {
		slog.Error("configuration error", "key", AWS_S3_ARN)
		os.Exit(1)
	}
	return _arn
}

func GetS3Bucket() string {
	_bucket := os.Getenv(AWS_S3_BUCKET)
	if _bucket == "" {
		slog.Error("configuration error", "key", AWS_S3_BUCKET)
		os.Exit(1)
	}
	return _bucket
}

func GetS3BucketKey() string {
	_bucketKey := os.Getenv(AWS_S3_BUCKET_KEY)
	if _bucketKey == "" {
		slog.Error("configuration error", "key", AWS_S3_BUCKET_KEY)
		os.Exit(1)
	}
	return _bucketKey
}

func GetBedrockDatasrouceID() string {
	datasourceID := os.Getenv(AWS_BEDROCK_DATASOURCE_ID)
	if datasourceID == "" {
		slog.Error("configuration error", "key", AWS_BEDROCK_DATASOURCE_ID)
	}
	return datasourceID
}

func GetBedrockKnowledgebaseID() string {
	knowledgeBaseID := os.Getenv(AWS_BEDROCK_KNOWLEDGEBASE_ID)
	if knowledgeBaseID == "" {
		slog.Error("configuration error", "key", AWS_BEDROCK_KNOWLEDGEBASE_ID)
	}
	return knowledgeBaseID
}

func GetBedrockAgentID() string {
	agentID := os.Getenv(AWS_BEDROCK_AGENT_ID)
	if agentID == "" {
		slog.Error("configuration error", "key", AWS_BEDROCK_AGENT_ID)
	}
	return agentID
}

func GetBedrockAgentAliasID() string {
	aliasID := os.Getenv(AWS_BEDROCK_AGENT_ALIAS_ID)
	if aliasID == "" {
		slog.Error("configuration error", "key", AWS_BEDROCK_AGENT_ALIAS_ID)
	}
	return aliasID
}

func GetLLModelARN() string {
	modelARN := strings.TrimSpace(os.Getenv(AWS_LL_MODEL_ARN))
	if modelARN == "" {
		slog.Error("configuration error", "key", AWS_LL_MODEL_ARN)
		os.Exit(1)
	}
	// if !strings.HasPrefix(modelARN, "arn:aws") {
	// 	slog.Error("configuration error", "key", AWS_LL_MODEL_ARN, "reason", "invalid ARN format")
	// 	os.Exit(1)
	// }
	return modelARN
}
