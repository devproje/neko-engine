package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Bot      BotConfig      `toml:"bot"`
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	Redis    RedisConfig    `toml:"redis"`
	Gemini   GeminiConfig   `toml:"gemini"`
	Memory   MemoryConfig   `toml:"memory"`
}

type BotConfig struct {
	ClientId         string `toml:"client-id"`
	ClientSecret     string `toml:"client-secret"`
	RedirectURI      string `toml:"redirect-uri"`
	OfficialServerId string `toml:"official-server-id"`
}

type ServerConfig struct {
	Host   string `toml:"host"`
	Port   int    `toml:"port"`
	Secret string `toml:"secret"`
}

type DatabaseConfig struct {
	URL      string `toml:"host"`
	Name     string `toml:"name"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type GeminiConfig struct {
	Token string `toml:"token"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	Database int    `toml:"database"`
}

type MemoryConfig struct {
	Model               string  `toml:"model"`
	ImportanceThreshold float64 `toml:"importance-threshold"`
	Enable              bool    `toml:"enable"`
}

type PromptConfig struct {
	Model   string `toml:"model"`
	Default string `toml:"default"`
	NSFW    string `toml:"nsfw"`
}

const (
	CONFIG_DEFAULT_BUF = `[bot]
client-id = "<discord client id>"
client-secret = "<discord client secret>"
redirect-uri = "<discord callback url>"
official-server-id = "<official discord server id>"

[server]
host = "0.0.0.0"
port = 3000

# please generate secret key and paste for gen-secret.sh script.
secret = ""

[memory]
model = "gemini-2.5-flash"
importance-threshold = 0.5
enable = true

[database]
host = "127.0.0.1"
port = 3306
name = "neko-engine"
username = "<mariadb_username>"
password = "<mariadb_password>"

[redis]
host = "127.0.0.1"
port = 6379
password = ""
database = 0

[gemini]
token = ""
`
	MODEL_DEFAULT_BUF = `model = "gemini-2.5-pro"
default = "<general_prompt>"

# NSFW only prompt. If you set this variable to empty, it will automatically fallback to the default prompt.
nsfw = ""
`
)

var (
	Debug      bool
	ConfigPath string

	Version   string
	Branch    string
	Hash      string
	BuildTime string
	GoVersion string
	Channel   string
)

type VersionInfo struct {
	Version   string `json:"version"`
	Branch    string `json:"branch"`
	Hash      string `json:"hash"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Channel   string `json:"channel"`
}

func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Branch:    Branch,
		Hash:      Hash,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		Channel:   Channel,
	}
}

func init() {
	ConfigPath = "./neko-data"

	pathEnv := os.Getenv("NEKO_PATH")
	if pathEnv != "" {
		ConfigPath = fmt.Sprintf("%s/neko-data", pathEnv)
	}

	if _, err := os.ReadDir(ConfigPath); err != nil {
		err = os.Mkdir(ConfigPath, 0755)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}

func Load() *Config {
	buf, err := os.ReadFile(filepath.Join(ConfigPath, "config.toml"))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config.toml is not found!\n")
		_ = os.WriteFile(filepath.Join(ConfigPath, "config.toml"), []byte(CONFIG_DEFAULT_BUF), 0644)

		return nil
	}

	var config Config
	if err = toml.Unmarshal(buf, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}

	return &config
}
