package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devproje/neko-engine/config"
	"github.com/pelletier/go-toml/v2"
)

type PromptService struct{}

type NKFile struct {
	Model  string `toml:"model"`
	Prompt struct {
		Default string `toml:"default"`
		NSFW    string `toml:"nsfw"`
	} `toml:"prompt"`
}

func Read(persona string) (*NKFile, error) {
	filename := fmt.Sprintf("%s.nkfile", persona)
	raw, err := os.ReadFile(filepath.Join(config.ConfigPath, "prompt", filename))
	if err != nil {
		return nil, err
	}

	var ret NKFile
	if err = toml.Unmarshal(raw, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}
