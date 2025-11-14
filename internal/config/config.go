package config

import (
	"time"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Driver string `mapstructure:"driver" yaml:"driver"`
	DSN    string `mapstructure:"dsn" yaml:"dsn"`
}

type VKConfig struct {
	BaseURL     string        `mapstructure:"base_url" yaml:"base_url"`
	APIVersion  string        `mapstructure:"api_version" yaml:"api_version"`
	AccessToken string        `mapstructure:"access_token" yaml:"access_token"`
	Timeout     time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

type GigaChatConfig struct {
	BaseURL string        `mapstructure:"base_url" yaml:"base_url"`
	Token   string        `mapstructure:"token" yaml:"token"`
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr" yaml:"addr"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"db" yaml:"db"`
}

type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
}

type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret" yaml:"jwt_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl" yaml:"refresh_token_ttl"`
	CookieDomain       string        `mapstructure:"cookie_domain" yaml:"cookie_domain"`
	CookieSecure       bool          `mapstructure:"cookie_secure" yaml:"cookie_secure"`
	CookieHTTPOnly     bool          `mapstructure:"cookie_httponly" yaml:"cookie_httponly"`
	VKClientID         string        `mapstructure:"vk_client_id" yaml:"vk_client_id"`
	VKClientSecret     string        `mapstructure:"vk_client_secret" yaml:"vk_client_secret"`
	VKRedirectURL      string        `mapstructure:"vk_redirect_url" yaml:"vk_redirect_url"`
	PasswordHashCost   int           `mapstructure:"password_hash_cost" yaml:"password_hash_cost"`
	RefreshTokenCookie string        `mapstructure:"refresh_token_cookie" yaml:"refresh_token_cookie"`
	AccessTokenCookie  string        `mapstructure:"access_token_cookie" yaml:"access_token_cookie"`
}

type HTTPClientConfig struct {
	ClientTimeout time.Duration `mapstructure:"client_timeout" yaml:"client_timeout"`
}

type HTTPConfig struct {
	Host         string        `mapstructure:"host" yaml:"host"`
	Port         int           `mapstructure:"port" yaml:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

type TelemetryConfig struct {
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled"`
	ServiceName  string `mapstructure:"service_name" yaml:"service_name"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint" yaml:"otlp_endpoint"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
}

type Config struct {
	DB         DBConfig         `mapstructure:"db" yaml:"db"`
	VK         VKConfig         `mapstructure:"vk" yaml:"vk"`
	GigaChat   GigaChatConfig   `mapstructure:"gigachat" yaml:"gigachat"`
	Redis      RedisConfig      `mapstructure:"redis" yaml:"redis"`
	Minio      MinioConfig      `mapstructure:"minio" yaml:"minio"`
	Auth       AuthConfig       `mapstructure:"auth" yaml:"auth"`
	HTTP       HTTPConfig       `mapstructure:"http" yaml:"http"`
	HTTPClient HTTPClientConfig `mapstructure:"http_client" yaml:"http_client"`
	Logging    LoggingConfig    `mapstructure:"logging" yaml:"logging"`
	Telemetry  TelemetryConfig  `mapstructure:"telemetry" yaml:"telemetry"`
	Metrics    MetricsConfig    `mapstructure:"metrics" yaml:"metrics"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("INTEAM")
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.SetConfigType("yaml")
	}

	_ = v.ReadInConfig()

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http_client.client_timeout", "15s")
	v.SetDefault("logging.level", "info")
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "720h")
	v.SetDefault("auth.cookie_secure", true)
	v.SetDefault("auth.cookie_httponly", true)
	v.SetDefault("auth.refresh_token_cookie", "inteam_refresh")
	v.SetDefault("auth.access_token_cookie", "inteam_access")
	v.SetDefault("auth.password_hash_cost", 12)
	v.SetDefault("telemetry.enabled", false)
	v.SetDefault("telemetry.service_name", "inteam-backend")
	v.SetDefault("metrics.enabled", true)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
