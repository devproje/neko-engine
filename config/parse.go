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
	OpenAI   OpenAIConfig   `toml:"openai"`
	Gemini   GeminiConfig   `toml:"gemini"`
}

type BotConfig struct {
	Token              string `toml:"token"`
	ClientId           string `toml:"client-id"`
	ClientSecret       string `toml:"client-secret"`
	RedirectURI        string `toml:"redirect-uri"`
	OfficialServerId   string `toml:"official-server-id"`
	ExperminalServerId string `toml:"experminal-server-id"`
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

type OpenAIConfig struct {
	Token string `toml:"token"`
}

type GeminiConfig struct {
	Token string `toml:"token"`
}

type PromptConfig struct {
	Default string `toml:"default"`
	NSFW    string `toml:"nsfw"`
}

var (
	Debug      bool
	ConfigPath string
)

func init() {
	ConfigPath = "./neko-data"

	pathEnv := os.Getenv("NEKO_PATH")
	if pathEnv != "" {
		ConfigPath = fmt.Sprintf("%s/neko-data", pathEnv)
	}

	if _, err := os.ReadDir(ConfigPath); err != nil {
		err = os.Mkdir(ConfigPath, 0644)
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
		return nil
	}

	var config Config
	if err = toml.Unmarshal(buf, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}

	return &config
}

func LoadPrompt() *PromptConfig {
	buf, err := os.ReadFile(filepath.Join(ConfigPath, "prompt.toml"))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "prompt.toml is not found!\n")
		return nil
	}

	var prompt PromptConfig
	if err = toml.Unmarshal(buf, &prompt); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}

	return &prompt
}
