package configs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"edu-schedule-system/internal/pkg/env"

	"github.com/spf13/viper"
)

var (
	config     Config
	configOnce sync.Once
	configErr  error
)

type Config struct {
	MySQL struct {
		Read struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			Name string `toml:"name"`
		} `toml:"read"`
		Write struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			Name string `toml:"name"`
		} `toml:"write"`
		Pool struct {
			MaxOpenConns           int   `toml:"maxOpenConns"`
			MaxIdleConns           int   `toml:"maxIdleConns"`
			ConnMaxLifetimeSeconds int64 `toml:"connMaxLifetimeSeconds"`
		} `toml:"pool"`
	} `toml:"mysql"`

	Redis struct {
		Enabled bool   `toml:"enabled"`
		Addr    string `toml:"addr"`
		Pass    string `toml:"pass"`
		Db      int    `toml:"db"`
		Pool    struct {
			DialTimeoutSeconds  int64 `toml:"dialTimeoutSeconds"`
			ReadTimeoutSeconds  int64 `toml:"readTimeoutSeconds"`
			WriteTimeoutSeconds int64 `toml:"writeTimeoutSeconds"`
			PoolSize            int   `toml:"poolSize"`
			MinIdleConns        int   `toml:"minIdleConns"`
			PoolTimeoutSeconds  int64 `toml:"poolTimeoutSeconds"`
		} `toml:"pool"`
	} `toml:"redis"`

	Auth struct {
		Mode                    string `toml:"mode"`
		TemplateTokenEnabled    bool   `toml:"templateTokenEnabled"`
		TemplateTokenConfigured bool   `toml:"-" mapstructure:"-"`
		RecoveryDelivery        struct {
			Mode           string `toml:"mode"`
			TimeoutSeconds int64  `toml:"timeoutSeconds"`
			MaxAttempts    int    `toml:"maxAttempts"`
			Email          struct {
				Host      string `toml:"host"`
				Port      int    `toml:"port"`
				Username  string `toml:"username"`
				Password  string `toml:"password"`
				From      string `toml:"from"`
				LocalName string `toml:"localName"`
			} `toml:"email"`
			SMS struct {
				Provider string `toml:"provider"`
				SignName string `toml:"signName"`
			} `toml:"sms"`
		} `toml:"recoveryDelivery"`
	} `toml:"auth"`

	Mongo struct {
		URI        string `toml:"uri"`
		UserName   string `toml:"username"`
		Password   string `toml:"password"`
		AuthSource string `toml:"authSource"`
	} `toml:"mongo"`

	JWT struct {
		Secret        string `toml:"secret"`
		Issuer        string `toml:"issuer"`
		Audience      string `toml:"audience"`
		LeewaySeconds int64  `toml:"leewaySeconds"`
	} `toml:"jwt"`

	AES struct {
		Secret string `toml:"secret"`
	} `toml:"aes"`

	RSA struct {
		PublicKey  string `toml:"publicKey"`
		PrivateKey string `toml:"privateKey"`
	} `toml:"rsa"`

	Language struct {
		Local string `toml:"local"`
	} `toml:"language"`

	Server struct {
		Port                     string `toml:"port"`
		MaxBodyBytes             int64  `toml:"maxBodyBytes"`
		ShutdownTimeoutSeconds   int64  `toml:"shutdownTimeoutSeconds"`
		ReadHeaderTimeoutSeconds int64  `toml:"readHeaderTimeoutSeconds"`
		ReadTimeoutSeconds       int64  `toml:"readTimeoutSeconds"`
		WriteTimeoutSeconds      int64  `toml:"writeTimeoutSeconds"`
		IdleTimeoutSeconds       int64  `toml:"idleTimeoutSeconds"`
	} `toml:"server"`
}

func init() {
	flag.String("config", "", "config file path (toml)")
}

func Init() error {
	loadConfig()
	return configErr
}

func MustInit() {
	if err := Init(); err != nil {
		panic(err)
	}
}

func Get() Config {
	if err := Init(); err != nil {
		panic(err)
	}
	return config
}

func Load() (Config, error) {
	if err := Init(); err != nil {
		return config, err
	}
	return config, nil
}

func loadConfig() {
	configOnce.Do(func() {
		cfgPath := resolveConfigPath()
		v := viper.New()
		v.SetConfigFile(cfgPath)
		v.SetConfigType("toml")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		if err := v.ReadInConfig(); err != nil {
			configErr = err
			return
		}

		setDefaults(v)

		templateTokenConfigured := v.InConfig("auth.templateTokenEnabled") || envIsSet("AUTH_TEMPLATE_TOKEN_ENABLED")

		if err := v.Unmarshal(&config); err != nil {
			configErr = err
			return
		}
		config.Auth.TemplateTokenConfigured = templateTokenConfigured

		if err := validateConfig(&config); err != nil {
			configErr = err
			return
		}
	})
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("redis.enabled", true)
	v.SetDefault("auth.mode", "auto")
	v.SetDefault("auth.recoveryDelivery.mode", "placeholder")
	v.SetDefault("auth.recoveryDelivery.timeoutSeconds", 10)
	v.SetDefault("auth.recoveryDelivery.maxAttempts", 2)
	v.SetDefault("auth.recoveryDelivery.email.port", 587)
	v.SetDefault("auth.recoveryDelivery.sms.provider", "stub")
	v.SetDefault("mysql.pool.maxOpenConns", 100)
	v.SetDefault("mysql.pool.maxIdleConns", 5)
	v.SetDefault("mysql.pool.connMaxLifetimeSeconds", 120)
	v.SetDefault("redis.pool.dialTimeoutSeconds", 5)
	v.SetDefault("redis.pool.readTimeoutSeconds", 20)
	v.SetDefault("redis.pool.writeTimeoutSeconds", 20)
	v.SetDefault("redis.pool.poolSize", 50)
	v.SetDefault("redis.pool.minIdleConns", 2)
	v.SetDefault("redis.pool.poolTimeoutSeconds", 60)
	v.SetDefault("server.readHeaderTimeoutSeconds", 5)
	v.SetDefault("server.readTimeoutSeconds", 10)
	v.SetDefault("server.writeTimeoutSeconds", 15)
	v.SetDefault("server.idleTimeoutSeconds", 60)
	_ = v.BindEnv("redis.enabled", "REDIS_ENABLED")
	_ = v.BindEnv("auth.mode", "AUTH_MODE")
	_ = v.BindEnv("auth.templateTokenEnabled", "AUTH_TEMPLATE_TOKEN_ENABLED")
	_ = v.BindEnv("auth.recoveryDelivery.mode", "AUTH_RECOVERY_DELIVERY_MODE")
	_ = v.BindEnv("auth.recoveryDelivery.timeoutSeconds", "AUTH_RECOVERY_DELIVERY_TIMEOUT_SECONDS")
	_ = v.BindEnv("auth.recoveryDelivery.maxAttempts", "AUTH_RECOVERY_DELIVERY_MAX_ATTEMPTS")
	_ = v.BindEnv("auth.recoveryDelivery.email.host", "AUTH_RECOVERY_DELIVERY_EMAIL_HOST")
	_ = v.BindEnv("auth.recoveryDelivery.email.port", "AUTH_RECOVERY_DELIVERY_EMAIL_PORT")
	_ = v.BindEnv("auth.recoveryDelivery.email.username", "AUTH_RECOVERY_DELIVERY_EMAIL_USERNAME")
	_ = v.BindEnv("auth.recoveryDelivery.email.password", "AUTH_RECOVERY_DELIVERY_EMAIL_PASSWORD")
	_ = v.BindEnv("auth.recoveryDelivery.email.from", "AUTH_RECOVERY_DELIVERY_EMAIL_FROM")
	_ = v.BindEnv("auth.recoveryDelivery.email.localName", "AUTH_RECOVERY_DELIVERY_EMAIL_LOCAL_NAME")
	_ = v.BindEnv("auth.recoveryDelivery.sms.provider", "AUTH_RECOVERY_DELIVERY_SMS_PROVIDER")
	_ = v.BindEnv("auth.recoveryDelivery.sms.signName", "AUTH_RECOVERY_DELIVERY_SMS_SIGN_NAME")
}

func envIsSet(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func resolveConfigPath() string {
	if v := strings.TrimSpace(configPathFromArgs(os.Args)); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("CONFIG_PATH")); v != "" {
		return v
	}
	envVal := env.Active().Value()
	return defaultConfigPath(envVal)
}

func defaultConfigPath(envVal string) string {
	name := filepath.Join("configs", fmt.Sprintf("%s_configs.toml", envVal))
	if _, err := os.Stat(name); err == nil {
		return name
	}

	dir, err := os.Getwd()
	if err != nil {
		return name
	}

	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return name
		}
		dir = parent
	}
}

func configPathFromArgs(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "-config" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "-config="))
		}
	}
	return ""
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return fmt.Errorf("config: jwt.secret is required")
	}
	if env.Active().IsPro() && weakJWTSecret(cfg.JWT.Secret) {
		return fmt.Errorf("config: jwt.secret must be a strong non-default value in pro")
	}
	if strings.TrimSpace(cfg.MySQL.Read.Addr) == "" {
		return fmt.Errorf("config: mysql.read.addr is required")
	}
	if strings.TrimSpace(cfg.MySQL.Read.User) == "" {
		return fmt.Errorf("config: mysql.read.user is required")
	}
	if strings.TrimSpace(cfg.MySQL.Read.Name) == "" {
		return fmt.Errorf("config: mysql.read.name is required")
	}
	if strings.TrimSpace(cfg.MySQL.Write.Addr) == "" {
		return fmt.Errorf("config: mysql.write.addr is required")
	}
	if strings.TrimSpace(cfg.MySQL.Write.User) == "" {
		return fmt.Errorf("config: mysql.write.user is required")
	}
	if strings.TrimSpace(cfg.MySQL.Write.Name) == "" {
		return fmt.Errorf("config: mysql.write.name is required")
	}
	if cfg.Redis.Enabled && strings.TrimSpace(cfg.Redis.Addr) == "" {
		return fmt.Errorf("config: redis.addr is required")
	}

	cfg.Auth.Mode = strings.ToLower(strings.TrimSpace(cfg.Auth.Mode))
	if cfg.Auth.Mode != "" && cfg.Auth.Mode != "auto" && cfg.Auth.Mode != "required" && cfg.Auth.Mode != "disabled" {
		return fmt.Errorf("config: auth.mode must be auto, required, or disabled")
	}
	if env.Active().IsPro() && cfg.Auth.Mode == "disabled" {
		return fmt.Errorf("config: auth.mode must not be disabled in pro")
	}
	if env.Active().IsPro() && cfg.Auth.TemplateTokenEnabled {
		return fmt.Errorf("config: auth.templateTokenEnabled must be false in pro")
	}
	cfg.Auth.RecoveryDelivery.Mode = strings.ToLower(strings.TrimSpace(cfg.Auth.RecoveryDelivery.Mode))
	if cfg.Auth.RecoveryDelivery.Mode == "" {
		cfg.Auth.RecoveryDelivery.Mode = "placeholder"
	}
	if cfg.Auth.RecoveryDelivery.Mode != "disabled" && cfg.Auth.RecoveryDelivery.Mode != "placeholder" && cfg.Auth.RecoveryDelivery.Mode != "email" && cfg.Auth.RecoveryDelivery.Mode != "sms" && cfg.Auth.RecoveryDelivery.Mode != "sms_stub" && cfg.Auth.RecoveryDelivery.Mode != "stub_sms" {
		return fmt.Errorf("config: auth.recoveryDelivery.mode must be disabled, placeholder, email, sms, sms_stub, or stub_sms")
	}
	if env.Active().IsPro() && cfg.Auth.RecoveryDelivery.Mode == "placeholder" {
		return fmt.Errorf("config: auth.recoveryDelivery.mode must not be placeholder in pro")
	}
	if cfg.Auth.RecoveryDelivery.Mode == "email" {
		if strings.TrimSpace(cfg.Auth.RecoveryDelivery.Email.Host) == "" {
			return fmt.Errorf("config: auth.recoveryDelivery.email.host is required when mode=email")
		}
		if cfg.Auth.RecoveryDelivery.Email.Port <= 0 {
			return fmt.Errorf("config: auth.recoveryDelivery.email.port must be positive when mode=email")
		}
		if strings.TrimSpace(cfg.Auth.RecoveryDelivery.Email.Username) == "" {
			return fmt.Errorf("config: auth.recoveryDelivery.email.username is required when mode=email")
		}
		if strings.TrimSpace(cfg.Auth.RecoveryDelivery.Email.Password) == "" {
			return fmt.Errorf("config: auth.recoveryDelivery.email.password is required when mode=email")
		}
		if strings.TrimSpace(cfg.Auth.RecoveryDelivery.Email.From) == "" {
			return fmt.Errorf("config: auth.recoveryDelivery.email.from is required when mode=email")
		}
	}
	if cfg.Auth.RecoveryDelivery.Mode == "sms" && strings.TrimSpace(cfg.Auth.RecoveryDelivery.SMS.Provider) == "" {
		return fmt.Errorf("config: auth.recoveryDelivery.sms.provider is required when mode=sms")
	}
	if cfg.Auth.RecoveryDelivery.TimeoutSeconds <= 0 {
		cfg.Auth.RecoveryDelivery.TimeoutSeconds = 10
	}
	if cfg.Auth.RecoveryDelivery.MaxAttempts <= 0 {
		cfg.Auth.RecoveryDelivery.MaxAttempts = 2
	}

	cfg.Server.Port = normalizePort(cfg.Server.Port)
	if cfg.Server.Port == "" {
		return fmt.Errorf("config: server.port is required")
	}
	if cfg.Server.ShutdownTimeoutSeconds <= 0 {
		cfg.Server.ShutdownTimeoutSeconds = 10
	}
	if cfg.Server.ReadHeaderTimeoutSeconds <= 0 {
		cfg.Server.ReadHeaderTimeoutSeconds = 5
	}
	if cfg.Server.ReadTimeoutSeconds <= 0 {
		cfg.Server.ReadTimeoutSeconds = 10
	}
	if cfg.Server.WriteTimeoutSeconds <= 0 {
		cfg.Server.WriteTimeoutSeconds = 15
	}
	if cfg.Server.IdleTimeoutSeconds <= 0 {
		cfg.Server.IdleTimeoutSeconds = 60
	}
	if cfg.Server.MaxBodyBytes < 0 {
		return fmt.Errorf("config: server.maxBodyBytes must be greater than or equal to 0")
	}
	if cfg.MySQL.Pool.MaxOpenConns <= 0 {
		cfg.MySQL.Pool.MaxOpenConns = 100
	}
	if cfg.MySQL.Pool.MaxIdleConns < 0 {
		return fmt.Errorf("config: mysql.pool.maxIdleConns must be greater than or equal to 0")
	}
	if cfg.MySQL.Pool.MaxIdleConns == 0 {
		cfg.MySQL.Pool.MaxIdleConns = 5
	}
	if cfg.MySQL.Pool.ConnMaxLifetimeSeconds <= 0 {
		cfg.MySQL.Pool.ConnMaxLifetimeSeconds = 120
	}
	if cfg.Redis.Pool.DialTimeoutSeconds <= 0 {
		cfg.Redis.Pool.DialTimeoutSeconds = 5
	}
	if cfg.Redis.Pool.ReadTimeoutSeconds <= 0 {
		cfg.Redis.Pool.ReadTimeoutSeconds = 20
	}
	if cfg.Redis.Pool.WriteTimeoutSeconds <= 0 {
		cfg.Redis.Pool.WriteTimeoutSeconds = 20
	}
	if cfg.Redis.Pool.PoolSize <= 0 {
		cfg.Redis.Pool.PoolSize = 50
	}
	if cfg.Redis.Pool.MinIdleConns < 0 {
		return fmt.Errorf("config: redis.pool.minIdleConns must be greater than or equal to 0")
	}
	if cfg.Redis.Pool.PoolTimeoutSeconds <= 0 {
		cfg.Redis.Pool.PoolTimeoutSeconds = 60
	}
	return nil
}

func weakJWTSecret(secret string) bool {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < 32 {
		return true
	}
	lower := strings.ToLower(trimmed)
	weakFragments := []string{
		"dev-secret",
		"local-only",
		"change-me",
		"changeme",
		"example",
		"template",
		"secret-local",
	}
	for _, fragment := range weakFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

func normalizePort(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return ""
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
