package constants

import "time"

const (
	ConfigPath           = "CONFIG_PATH"
	AppEnv               = "APP_ENV"
	AppRootPath          = "APP_ROOT_PATH"
	ProjectNameEnv       = "PROJECT_NAME"
	Json                 = "json"
	Grpc                 = "GRPC"
	Method               = "METHOD"
	Name                 = "NAME"
	Metadata             = "METADATA"
	Request              = "REQUEST"
	Reply                = "REPLY"
	Time                 = "TIME"
	MaxHeaderBytes       = 1 << 20
	StackSize            = 1 << 10 // 1 KB
	BodyLimit            = "2M"
	ReadTimeout          = 15 * time.Second
	WriteTimeout         = 15 * time.Second
	GzipLevel            = 5
	WaitShotDownDuration = 3 * time.Second
	Dev                  = "development"
	Test                 = "test"
	Production           = "production"
)

const (
	ErrBadRequestTitle          = "Bad Request"
	ErrConflictTitle            = "Conflict Error"
	ErrNotFoundTitle            = "Not Found"
	ErrUnauthorizedTitle        = "Unauthorized"
	ErrForbiddenTitle           = "Forbidden"
	ErrRequestTimeoutTitle      = "Request Timeout"
	ErrInternalServerErrorTitle = "Internal Server Error"
	ErrDomainTitle              = "Domain Model Error"
	ErrApplicationTitle         = "Application Service Error"
	ErrApiTitle                 = "Api Error"
)
