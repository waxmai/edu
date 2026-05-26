package notification

type Config struct {
	Mode           string
	TimeoutSeconds int64
	MaxAttempts    int
	Email          EmailConfig
	SMS            SMSConfig
}

type EmailConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	LocalName string
}

type SMSConfig struct {
	Provider string
	SignName string
}
