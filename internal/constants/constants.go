package constants

const (
	EnvProduction  = "production"
	EnvStaging     = "staging"
	EnvDevelopment = "development"
	EnvLocal       = "local"
)

const (
	DefaultAppKey  string = "repo-scanner"
	DefaultAppName string = "Repository Scanner"
	DefaultAppPort int    = 8080
)

const (
	AppDebug   = "APP_DEBUG"
	AppEnv     = "APP_ENV"
	AppKey     = "APP_KEY"
	AppName    = "APP_NAME"
	AppVersion = "APP_VERSION"
	AppHost    = "APP_HOST"
	AppPort    = "APP_PORT"

	DBEngine       = "DB_ENGINE"
	DBHost         = "DB_HOST"
	DBPort         = "DB_PORT"
	DBHostRW       = "DB_HOST_RW"
	DBPortRW       = "DB_PORT_RW"
	DBHostRO       = "DB_HOST_RO"
	DBPortRO       = "DB_PORT_RO"
	DBUser         = "DB_USER"
	DBPwd          = "DB_PWD"
	DBName         = "DB_NAME"
	DBSSLMode      = "DB_SSL_MODE"
	DBConnStr      = "DB_CONN_STR"
	DBConnLifetime = "DB_CONN_LIFETIME"
	DBConnMaxIdle  = "DB_CONN_MAX_IDLE"
	DBConnMaxOpen  = "DB_CONN_MAX_OPEN"
)

const (
	DefaultTimezone                = "+7"
	RegexUnnecessaryInquiryDetails = `^rp`
	DefaultLimit                   = 10
	DefaultPage                    = 1
)

const (
	ScanningStatusQueued     = "queued"
	ScanningStatusInProgress = "in_progress"
	ScanningStatusSuccess    = "success"
	ScanningStatusFailure    = "failure"
)

var ScanningInProgress bool
