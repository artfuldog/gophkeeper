package server

import (
	"fmt"
	"log"

	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/server/db"
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

// Default configuration parameters.
const (
	defAddress  = "127.0.0.1:3200"
	defDBType   = db.TypePostgres
	defSyncType = "postgres"
)

//nolint:gochecknoglobals
var (
	defLogLevel         = fmt.Sprint(logger.WarnLevel)
	defMaxSecretSize    = uint32(50 * 1024 * 1024) // 50 Mb
	defTokenValidPeriod = uint32(30 * 60)          // 30 minutes
)

// Config represents server's configurations parameters.
type Config struct {
	// Server address.
	// Supported format: <ip-address/fqdn/hostname>:<port>, ex. 10.20.30.40:3200, my.host.com:3333
	Address string `env:"GK_ADDRESS"`

	// Database type (postrgres).
	DBType string `env:"GK_DB_TYPE"`
	// Database dsn in format address:port/db_name.
	DBDSN string `env:"GK_DB_DSN"`
	// Database user.
	DBUser string `env:"GK_DB_USER"`
	// Database user.
	DBPassword string `env:"GK_DB_PASSWORD"`

	// TLS Certificate file (.pem).
	TLSCertFilepath string `env:"GK_TLS_CERT"`
	// TLS Certificate key file (.key).
	TLSKeyFilepath string `env:"GK_TLS_KEY"`
	// Disable TLS encryption.
	TLSDisable bool

	// Log level (debug/info/warn/error/fatal/panic).
	LogLevel string `env:"GK_LOG_LEVEL"`
	// Maximum secret size in bytes.
	MaxSecretSize uint32 `env:"GK_MAX_SECRET"`
	// Server key. Used for generated tokens. Must be 32-byte length.
	ServerKey string `env:"GK_SERVER_KEY"`
	// Token valid period in seconds.
	TokenValidPeriod uint32 `env:"GK_TOKEN_EXP"`
}

// NewConfig a helper function for reading cli arguments and environmental variables and
// prepare server configuration.
//
// Envvars have more priority than cli arguments.
func NewConfig() (*Config, error) {
	cfg := new(Config)

	// Read cli arguments
	flag.StringVarP(&cfg.Address, "address", "a", defAddress, "address and port of server in format ip:port")

	flag.StringVarP(&cfg.DBType, "dbtype", "D", defDBType, "database type (postrgres)")
	flag.StringVarP(&cfg.DBDSN, "dbdsn", "d", "", "database dsn in format address:port")
	flag.StringVar(&cfg.DBUser, "db_user", "", "database user (should be set via cli only for testing)")
	flag.StringVar(&cfg.DBPassword, "db_password", "", "database password (should be set via cli only for testing)")

	flag.StringVar(&cfg.TLSCertFilepath, "tls-cert", "", "path to TLS certificate file (.pem)")
	flag.StringVar(&cfg.TLSKeyFilepath, "tls-key", "", "path to TLS Certificate key file (.key)")
	flag.BoolVar(&cfg.TLSDisable, "disable-tls", false, "disable TLS")

	flag.StringVarP(&cfg.LogLevel, "loglevel", "l", defLogLevel, "log level (debug/info/warn/error/fatal/panic)")
	flag.Uint32VarP(&cfg.MaxSecretSize, "max_size", "m", defMaxSecretSize, "maximum secret size in bytes")
	flag.StringVarP(&cfg.ServerKey, "server_key", "k", "", "server key(should be set via cli only for testing)")
	flag.Uint32VarP(&cfg.TokenValidPeriod, "token_exp", "t", defTokenValidPeriod, "token valid period in seconds")

	flag.Parse()

	// Read environment variables
	if err := env.Parse(cfg); err != nil {
		log.Print("failed to read env vars")
	}

	return cfg, nil
}
