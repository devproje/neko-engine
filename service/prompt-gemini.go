package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/model"
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
				{GoogleSearch: &genai.GoogleSearch{}},
			},
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

func (*GeminiService) AbstractDataFromLLM(acc *model.Account, prompt []*genai.Content, nsfw bool) (*Memory, int, error) {
	cnf := config.Load()
	persona := config.LoadPrompt()
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  cnf.Gemini.Token,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Printf("[MemoryService] 기억저장 거부\n")
		return nil, 0, err
	}

	ret := persona.Default
	if nsfw {
		if persona.NSFW != "" {
			ret = persona.NSFW
		}
	}

	system := "The input below contains the user's core information.\n\n"
	system += "Convert this sentence into a one-line JSON object in the following format:\n"
	system += "{ \"key\": \"keyword\", \"content\": \"sentence\", \"importance\": 5, \"sentiment_score\": 0 }\n\n"
	system += "- For key, specify a single Korean word representing the main attribute of the sentence (e.g., hobby, habit, emotion, preference). (varchar type, up to 40 characters)\n"
	system += "- For content, generate a 'memory' sentence summarizing the meaning. Example: The user enjoys watching parrots as a hobby.\n"
	system += "- For importance, rate the importance as an integer from 0 to 10.\n"
	system += "- For sentiment_score, analyze the conversation tone and user-bot interaction to determine sentiment change from -25 (extremely negative) to 15 (extremely positive). Consider: user's attitude, bot's response quality, emotional context, and overall interaction flow.\n"
	system += "- Output only a single, accurate JSON object. Comments, explanations, arrays, or multi-line output are strictly prohibited.\n\n"
	system += "- If any of the following conditions apply, do not store memory but keep sentiment_score:\n"
	system += "  1) The input contains profanity.\n"
	system += "  2) The input appears to simply repeat the chatbot's response.\n"
	system += "  3) The chatbot's response is negative, uncooperative, or not worth remembering.\n"
	system += "- You must always consider both the user's input and the chatbot's answer when making a judgment.\n\n"
	system += "- You must strictly follow these requirements; if any are violated, the system is considered to have failed:\n"
	system += "  * Only a single line of JSON output is allowed.\n"
	system += "  * CRITICAL: sentiment_score field is MANDATORY and must ALWAYS be included in the JSON response.\n"
	system += "  * If memory is not worth saving, return empty key and content but ALWAYS include sentiment_score.\n\n"
	system += "Example:\n"
	system += "{ \"key\": \"취미\", \"content\": \"사용자는 앵무새를 보는것이 취미이다.\", \"importance\": 5, \"sentiment_score\": 3 }\n"
	system += fmt.Sprintf("\n\nThis is system persona:\n%s\n", ret)

	responseSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"key": {
				Type:        genai.TypeString,
				Description: "A single Korean word representing the main attribute of the sentence (e.g., hobby, habit, emotion, preference).",
				MaxLength:   func() *int64 { v := int64(40); return &v }(),
			},
			"content": {
				Type:        genai.TypeString,
				Description: "A memory sentence summarizing the meaning.",
			},
			"importance": {
				Type:        genai.TypeInteger,
				Description: "Importance as an integer from 0 to 10.",
				Minimum:     func() *float64 { v := float64(0); return &v }(),
				Maximum:     func() *float64 { v := float64(10); return &v }(),
			},
			"sentiment_score": {
				Type:        genai.TypeInteger,
				Description: "Sentiment score of the memory, from -25 (extremely negative) to 15 (extremely positive). Wide range for dramatic reactions.",
				Minimum:     func() *float64 { v := float64(-25); return &v }(),
				Maximum:     func() *float64 { v := float64(15); return &v }(),
			},
		},
		Required: []string{"key", "content", "importance", "sentiment_score"},
	}

	answer, err := client.Models.GenerateContent(
		context.Background(),
		"gemini-2.5-flash-preview-05-20",
		prompt,
		&genai.GenerateContentConfig{
			ResponseSchema: responseSchema,
			SystemInstruction: &genai.Content{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					genai.NewPartFromText(system),
				},
			},
		},
	)
	if err != nil {
		return nil, 0, err
	}

	fmt.Println(answer.Text())

	// Extract JSON object from response (only content within {})
	jsonStart := strings.Index(answer.Text(), "{")
	jsonEnd := strings.LastIndex(answer.Text(), "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		return nil, 0, fmt.Errorf("no valid JSON object found in response")
	}
	jsonStr := answer.Text()[jsonStart : jsonEnd+1]

	var parsed struct {
		MemKey         string `json:"key"`
		Content        string `json:"content"`
		Importance     int    `json:"importance"`
		SentimentScore int    `json:"sentiment_score"`
	}
	if err = json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, 0, err
	}

	if parsed.MemKey == "" || parsed.Content == "" {
		fmt.Println("[MemoryService] 기억저장 거부")
		return nil, parsed.SentimentScore, nil
	}

	var memory = &Memory{
		UserID:     acc.Id,
		MemKey:     parsed.MemKey,
		Content:    parsed.Content,
		Importance: parsed.Importance,
	}

	fmt.Printf("[MemoryService] 사용자가 기억저장을 요청함: [%s] %s (중요도: %d)\ns", memory.MemKey, memory.Content, memory.Importance)
	return memory, parsed.SentimentScore, nil
}
