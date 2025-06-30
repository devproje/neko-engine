package common

import (
	"fmt"
	"os"

	"github.com/devproje/neko-engine/common/controller"
	"github.com/devproje/neko-engine/common/service"
	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/util"
)

type ServiceLoader struct {
	Acc     *controller.AccountController
	Chat    *controller.ChatController
	Admin   *controller.AdminController
	Memory  *controller.MemoryController
	Role    *controller.RoleController
	Stats   *controller.StatsController
	Discord *controller.DiscordController
	Account *service.AccountService
	Gemini  *service.GeminiService
	Prompt  *service.PromptService
	MemServ *service.MemoryService
}

func New() *ServiceLoader {
	cfg := config.Load()
	if cfg != nil {
		if err := util.InitializeRedis(&cfg.Redis); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Redis: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Redis connection established")
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load config, Redis not initialized\n")
	}

	account := service.NewAccountService()
	gemini := service.NewGeminiService()
	memory := service.NewMemoryService()
	prompt := service.NewPromptService()

	acc := controller.NewAccountController(account)
	chat := controller.NewChatController(account, gemini, memory, prompt)
	admin := controller.NewAdminController(account)
	memCtrl := controller.NewMemoryController(memory)
	role := controller.NewRoleController(account)
	stats := controller.NewStatsController(account, memory)
	discord := controller.NewDiscordController(account)

	return &ServiceLoader{
		Acc:     acc,
		Chat:    chat,
		Admin:   admin,
		Memory:  memCtrl,
		Role:    role,
		Stats:   stats,
		Discord: discord,
		Account: account,
		Prompt:  prompt,
		Gemini:  gemini,
		MemServ: memory,
	}
}
