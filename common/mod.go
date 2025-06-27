package common

import (
	"github.com/devproje/neko-engine/common/controller"
	"github.com/devproje/neko-engine/common/service"
)

type ServiceLoader struct {
	Acc     *controller.AccountController
	Chat    *controller.ChatController
	Account *service.AccountService
	Gemini  *service.GeminiService
	Prompt  *service.PromptService
}

func New() *ServiceLoader {
	account := service.NewAccountService()
	gemini := service.NewGeminiService()
	memory := service.NewMemoryService()
	prompt := service.NewPromptService()

	acc := controller.NewAccountController(account)
	chat := controller.NewChatController(account, gemini, memory, prompt)

	return &ServiceLoader{
		Acc:     acc,
		Chat:    chat,
		Account: account,
		Prompt:  prompt,
		Gemini:  gemini,
	}
}
