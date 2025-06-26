package controller

import (
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type ChatController struct {
	Gemini *service.GeminiService
	Prompt *service.PromptService
}

type ChatForm struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Info   struct {
		Content string `json:"chat"`
		NSFW    bool   `json:"nsfw"`
	} `json:"info"`
}

func NewChatController(gemini *service.GeminiService, prompt *service.PromptService) *ChatController {
	return &ChatController{
		Gemini: gemini,
		Prompt: prompt,
	}
}

func (*ChatController) SendChat(ctx *gin.Context) {

}
