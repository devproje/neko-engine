package service

import (
	"context"

	"github.com/devproje/neko-engine/config"
	"google.golang.org/genai"
)

type GeminiService struct{}

func NewGeminiService() *GeminiService {
	return &GeminiService{}
}

func (*GeminiService) SendPrompt(system, model string, prompts []*genai.Content) (*genai.GenerateContentResponse, error) {
	cnf := config.Load()
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  cnf.Gemini.Token,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	system += "\noutput text length must be fewer 2000\n"
	result, err := client.Models.GenerateContent(
		context.Background(),
		model,
		prompts,
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					{Text: system},
				},
			},
			Tools: []*genai.Tool{
				{
					GoogleSearch: &genai.GoogleSearch{},
					URLContext:   &genai.URLContext{},
				},
			},
			Temperature: func() *float32 {
				var ret float32 = 0.5
				return &ret
			}(),
			MaxOutputTokens: 15000,
			ThinkingConfig: &genai.ThinkingConfig{
				IncludeThoughts: true,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
