package config

import (
	"fmt"
	"os"

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

var (
	Debug bool
)

func Load() *Config {
	buf, err := os.ReadFile("config.toml")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config.toml is not created!\n")
		return nil
	}

	var config Config
	if err = toml.Unmarshal(buf, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}

	return &config
}
