// Package config implemets configuration management interface for client.
//
// Configer, which represent conrolling entity, is concurency-safe for all read-write
// opertaions with config parameters.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/mitchellh/mapstructure"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Default configuration's file parameters.
const (
	appConfigName = "config.yaml"
	appConfigDir  = ".gophkeeper/"
)

// Errors.
var (
	ErrConfigNotFound     = errors.New("config file missed")
	ErrCreateAppDirFailed = errors.New("failed to create working directory")
	ErrEmptyUser          = errors.New("user is empty, must be set")
	ErrEmptySecretKey     = errors.New("secret key is empty, must be set")
	ErrEmptyServer        = errors.New("server address is empty, must be set")
	ErrEmptyAgentMode     = errors.New("agent mode is not set")
	ErrWrongSyncInterval  = errors.New("sync interval must be between 10 and 1800 seconds")
)

// Agent's modes.
type AgentMode int8

const (
	// Unknown mode.
	ModeUnknown AgentMode = iota
	// Server stores all items.
	ModeServer
	// Agent use local storage and peridically sync with Server (not implemented yet).
	ModeLocal
)

// UnmarshalText implements encoding.TextUnmarshaler interface and
// using a Viper's decode hook.
func (a *AgentMode) UnmarshalText(text []byte) error {
	switch string(text) {
	case "server":
		*a = ModeServer
	case "local":
		*a = ModeLocal
	default:
		*a = ModeUnknown
	}

	return nil
}

// String implements Stringer interface.
func (a AgentMode) String() string {
	switch a {
	case ModeServer:
		return "server"
	case ModeLocal:
		return "local"
	default:
		return "uknown mode"
	}
}

// Flags represent cli arguments (flags).
type Flags struct {
	ShowVersion      bool
	LogLevel         string
	CustomConfigPath string
	DisableTLS       bool
}

// ReadFlags reads cli arguments passed to agent.
func ReadFlags() *Flags {
	f := new(Flags)

	flag.BoolVarP(&f.ShowVersion, "version", "v", false, "show version")
	flag.StringVarP(&f.CustomConfigPath, "config", "c", "", "path to configuration file")
	flag.StringVarP(&f.LogLevel, "log_level", "l", "", "log level")
	flag.BoolVarP(&f.DisableTLS, "disable-tls", "t", false, "disable TLS (should be used only for testing)")
	flag.Parse()

	return f
}

// Configer represents entity which stores and controls configuration parameters.
// Based on Viper package.
//
// Supported configuration parameters
//   - User - username
//   - SecretKey - secret key, used for decrypt encryption key from server
//   - Server - server address in format <ip-address/fqdn/hostname>:<port>, ex. 10.20.30.40:3200, my.host.com:3333
//   - Mode - agent working mode - local / server
//   - SyncInterval - interval between synchronization with server in seconds (10 - 1800)
//   - ShowSensitive - show by default sensitive information in UI
//   - LogLevel - agent log level, currently useless, ignore it
//   - CAcert - path to CA root certificate. Recommended way - not use this optioin and install CA into system.
//   - Disable TLS - disables TLS encryption. Should be used only for testing/lab environments.
type Configer struct {
	*viper.Viper
	mu sync.RWMutex
}

// NewConfiger creates Configer instance.
func NewConfiger(flags *Flags) (*Configer, error) {
	cfg := &Configer{Viper: viper.New()}

	cfg.SetDefault("mode", ModeServer)
	cfg.SetDefault("syncinterval", 30*time.Second)
	cfg.SetDefault("showsensitive", false)
	cfg.SetDefault("loglevel", fmt.Sprint(logger.ErrorLevel))
	cfg.SetDefault("disabletls", false)

	if flags != nil && flags.CustomConfigPath != "" {
		cfg.SetConfigFile(flags.CustomConfigPath)
	} else {
		cfg.SetConfigName(appConfigName)
		cfg.AddConfigPath("./" + appConfigDir)
		cfg.AddConfigPath("$HOME/" + appConfigDir)
	}

	cfg.SetConfigType("yaml")

	if flags != nil {
		cfg.SetTLSDisable(flags.DisableTLS)
	}

	if err := cfg.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return cfg, ErrConfigNotFound
		}

		if errors.Is(err, fs.ErrNotExist) {
			return cfg, ErrConfigNotFound
		}

		return nil, err
	}

	return cfg, nil
}

// Validate validates current config.
//
// Returns nil on success.
func (c *Configer) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.IsSet("user") || c.GetString("user") == "" {
		return ErrEmptyUser
	}

	if !c.IsSet("secretkey") || c.GetString("secretkey") == "" {
		return ErrEmptySecretKey
	}

	if !c.IsSet("server") || c.GetString("server") == "" {
		return ErrEmptyServer
	}

	if !c.IsSet("mode") || c.unmarshallAgentMode() == ModeUnknown {
		return ErrEmptyAgentMode
	}

	if c.GetSyncInterval() < (10*time.Second) || c.GetSyncInterval() > (1800*time.Second) {
		return ErrWrongSyncInterval
	}

	return nil
}

// GetUser returns current user from config.
func (c *Configer) GetUser() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetString("user")
}

// SetUser sets username parameter.
func (c *Configer) SetUser(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("user", v)
}

// GetSecretKey returns current secret key.
func (c *Configer) GetSecretKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetString("secretkey")
}

// SetUser sets secret key parameter.
func (c *Configer) SetSecretKey(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("secretkey", v)
}

// GetServer returns current server address.
func (c *Configer) GetServer() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetString("server")
}

// SetServer sets server parameter.
func (c *Configer) SetServer(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("server", v)
}

// GetMode returns current agent working mode.
func (c *Configer) GetMode() AgentMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.unmarshallAgentMode()
}

// unmarshallAgentMode is a concurency-UNSAFE helper function for unmarshall Agent Mode.
func (c *Configer) unmarshallAgentMode() AgentMode {
	var mode AgentMode
	if err := c.UnmarshalKey("mode", &mode,
		viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return ModeUnknown
	}

	if mode < 1 || mode > 2 {
		return ModeUnknown
	}

	return mode
}

// SetMode sets agent working mode.
func (c *Configer) SetMode(v AgentMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("mode", fmt.Sprint(v))
}

// GetMode returns current sync interval.
func (c *Configer) GetSyncInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetDuration("syncinterval")
}

// SetMode sets current sync interval.
func (c *Configer) SetSyncInterval(v time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("syncinterval", v)
}

// GetShowSensitive returns current status of showing sensitive information.
func (c *Configer) GetShowSensitive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetBool("showsensitive")
}

// SetShowSensitive sets show sensitive information parameter.
func (c *Configer) SetShowSensitive(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("showsensitive", v)
}

// GetLogLevel returns current log level.
func (c *Configer) GetLogLevel() logger.Level {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.Get("loglevel").(string)
	if !ok {
		return logger.NoLevel
	}

	level, err := logger.GetLevelFromString(val)
	if err != nil {
		return logger.NoLevel
	}

	return level
}

// SetLogLevel sets current log level.
func (c *Configer) SetLogLevel(v logger.Level) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("loglevel", fmt.Sprint(v))
}

// GetCAcert returns current status of showing sensitive information.
func (c *Configer) GetCACert() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetString("cacert")
}

// SetShowSensitive sets show sensitive information parameter.
func (c *Configer) SetCACert(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("cacert", v)
}

// GetCAcert returns current TLS Disable flag value.
func (c *Configer) GetTLSDisable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetBool("disabletls")
}

// SetShowSensitive sets TLS Disable flag value.
func (c *Configer) SetTLSDisable(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set("disabletls", v)
}

// createAppDir creates agent directory.
func (c *Configer) CreateAppDir() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := os.Stat(appConfigDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(appConfigDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCreateAppDirFailed, err)
		}
	}

	return nil
}

// GetConfigDefaultFilepath returns default config path.
func (c *Configer) GetConfigDefaultFilepath() string {
	return (appConfigDir + appConfigName)
}

// GetAppConfigDir returns application config directory.
func (c *Configer) GetAppConfigDir() string {
	return appConfigDir
}
