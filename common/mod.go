package common

import (
	"github.com/devproje/neko-engine/common/controller"
	"github.com/devproje/neko-engine/common/service"
)

type ServiceLoader struct {
	Chat   *controller.ChatController
	Prompt *service.PromptService
	Gemini *service.GeminiService
}

func New() *ServiceLoader {
	gemini := service.NewGeminiService()
	prompt := service.NewPromptService()

	chat := controller.NewChatController(gemini, prompt)

	return &ServiceLoader{
		Chat:   chat,
		Prompt: prompt,
		Gemini: gemini,
	}
}
