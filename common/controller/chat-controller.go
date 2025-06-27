package controller

import (
	"fmt"

	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ChatController struct {
	Gemini *service.GeminiService
	Prompt *service.PromptService
}

type ChatForm struct {
	Id          string       `json:"id"`
	Author      string       `json:"author"`
	Content     string       `json:"content"`
	Persona     string       `json:"persona"`
	Attachments []Attachment `json:"attachments"`
	Info        struct {
		Content string `json:"chat"`
		NSFW    bool   `json:"nsfw"`
	} `json:"info"`
}

type Attachment struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
}

func NewChatController(gemini *service.GeminiService, prompt *service.PromptService) *ChatController {
	return &ChatController{
		Gemini: gemini,
		Prompt: prompt,
	}
}

func (cc *ChatController) SendChat(ctx *gin.Context) {
	// cnf := config.Load()
	var req ChatForm

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "some required parameter is not contained",
		})
		return
	}

	persona, err := cc.Prompt.Read(req.Persona)
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": fmt.Sprintf("'%s' persona is not found", req.Persona),
		})
		return
	}

	prompt := persona.Prompt.Default
	if req.Info.NSFW && persona.Prompt.NSFW != "" {
		prompt = persona.Prompt.NSFW
	}

	input := make([]*genai.Content, 0)
	input = append(input, genai.NewContentFromText(req.Content, genai.RoleUser))

	resp, err := cc.Gemini.SendPrompt(prompt, persona.Model, input)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Gemini API is not responding",
		})
	}

	ctx.JSON(200, gin.H{
		"answer": resp.Text(),
		"usage": gin.H{
			"prompt":    resp.UsageMetadata.PromptTokenCount,
			"candidate": resp.UsageMetadata.CandidatesTokenCount,
			"total":     resp.UsageMetadata.TotalTokenCount,
		},
	})
}
