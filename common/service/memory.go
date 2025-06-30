package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/util"
	"google.golang.org/genai"
)

type MemoryService struct{}

type MemoryData struct {
	UID       string                `json:"user_id"`
	Histories []*repository.History `json:"histories"`
	Memories  []*repository.Memory  `json:"memories"`
}

type ImportanceAnalysis struct {
	Importance float64  `json:"importance"`
	Summary    string   `json:"summary"`
	Reason     string   `json:"reason"`
	Keywords   []string `json:"keywords"`
}

func NewMemoryService() *MemoryService {
	return &MemoryService{}
}

func init() {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer db.Close()

	if err := db.GetDB().AutoMigrate(&repository.History{}, &repository.Memory{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

func (*MemoryService) LoadHistory(uid string) (*MemoryData, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	hist := repository.NewHistoryRepository(db)
	history, err := hist.Read(uid, 20) // load last chats
	if err != nil {
		return nil, err
	}

	md := MemoryData{
		UID:       uid,
		Histories: history,
	}

	return &md, nil
}

func (*MemoryService) AppendHistory(history *repository.History) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	hist := repository.NewHistoryRepository(db)
	if err := hist.Create(history); err != nil {
		return err
	}

	return nil
}

func (*MemoryService) PurgeLast(uid string) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	hist := repository.NewHistoryRepository(db)
	if err := hist.PurgeOne(uid); err != nil {
		return err
	}

	return nil
}

func (*MemoryService) PurgeN(uid string, n int) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	hist := repository.NewHistoryRepository(db)
	if err := hist.PurgeN(uid, n); err != nil {
		return err
	}

	return nil
}

func (*MemoryService) FlushHistory(uid string) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	hist := repository.NewHistoryRepository(db)
	if err := hist.Flush(uid); err != nil {
		return err
	}

	return nil
}

func (*MemoryService) AnalyzeImportance(userMessage, botMessage string) (*ImportanceAnalysis, error) {
	gemini := NewGeminiService()
	
	systemPrompt := `You are a memory importance analyzer. Analyze the conversation between user and bot and determine:
1. Importance score (0.0-1.0): How important is this conversation for future reference?
2. Summary: Brief summary of the key information
3. Reason: Why this conversation is important or not
4. Keywords: Array of 3-7 relevant keywords/phrases for searching this memory

Consider these factors:
- Personal information shared
- Preferences expressed
- Important decisions made
- Recurring topics
- Emotional context
- Technical knowledge shared

Keywords should include:
- Key topics discussed
- Specific names, places, things mentioned
- Skills or interests mentioned
- Important concepts or themes

Return JSON format: {"importance": 0.0-1.0, "summary": "brief summary", "reason": "explanation", "keywords": ["keyword1", "keyword2", "keyword3"]}`

	prompts := []*genai.Content{
		{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: fmt.Sprintf("User: %s\nBot: %s", userMessage, botMessage)},
			},
		},
	}

	response, err := gemini.SendPrompt(systemPrompt, "gemini-1.5-flash", prompts)
	if err != nil {
		return nil, err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	responseText := response.Candidates[0].Content.Parts[0].Text
	responseText = strings.TrimSpace(responseText)
	if after, found := strings.CutPrefix(responseText, "```json"); found {
		responseText = strings.TrimSuffix(after, "```")
		responseText = strings.TrimSpace(responseText)
	}

	var analysis ImportanceAnalysis
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v", err)
	}

	if analysis.Importance < 0.0 {
		analysis.Importance = 0.0
	} else if analysis.Importance > 1.0 {
		analysis.Importance = 1.0
	}

	return &analysis, nil
}

func (*MemoryService) SaveMemoryIfImportant(uid, userMessage, botMessage, providerID, providerUsername string) error {
	ms := &MemoryService{}
	analysis, err := ms.AnalyzeImportance(userMessage, botMessage)
	if err != nil {
		return err
	}

	if analysis.Importance < 0.5 {
		return nil
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	keywordsStr := strings.Join(analysis.Keywords, ",")
	memory := &repository.Memory{
		UserID:           uid,
		UserMessage:      userMessage,
		BotMessage:       botMessage,
		Importance:       analysis.Importance,
		Summary:          analysis.Summary,
		Keywords:         keywordsStr,
		ProviderID:       providerID,
		ProviderUsername: providerUsername,
	}

	return memRepo.Create(memory)
}

func (*MemoryService) LoadMemories(uid string, limit int) ([]*repository.Memory, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	return memRepo.ReadByImportance(uid, 0.5, limit)
}

func (*MemoryService) LoadMemoryData(uid string) (*MemoryData, error) {
	ms := &MemoryService{}
	
	historyData, err := ms.LoadHistory(uid)
	if err != nil {
		return nil, err
	}

	memories, err := ms.LoadMemories(uid, 10)
	if err != nil {
		return nil, err
	}

	return &MemoryData{
		UID:       uid,
		Histories: historyData.Histories,
		Memories:  memories,
	}, nil
}

func (*MemoryService) ExtractKeywords(text string) ([]string, error) {
	gemini := NewGeminiService()
	
	systemPrompt := `Extract 3-7 important keywords or key phrases from the given text. 
Focus on:
- Main topics or subjects
- Specific names, places, or things
- Important concepts or skills
- Action words or activities

Return only a JSON array of strings: ["keyword1", "keyword2", "keyword3"]`

	prompts := []*genai.Content{
		{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: text},
			},
		},
	}

	response, err := gemini.SendPrompt(systemPrompt, "gemini-1.5-flash", prompts)
	if err != nil {
		return nil, err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	responseText := response.Candidates[0].Content.Parts[0].Text
	responseText = strings.TrimSpace(responseText)
	if after, found := strings.CutPrefix(responseText, "```json"); found {
		responseText = strings.TrimSuffix(after, "```")
		responseText = strings.TrimSpace(responseText)
	}

	var keywords []string
	if err := json.Unmarshal([]byte(responseText), &keywords); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v", err)
	}

	return keywords, nil
}

func (*MemoryService) SearchMemoriesByKeywords(uid, query string, limit int) ([]*repository.Memory, error) {
	ms := &MemoryService{}
	keywords, err := ms.ExtractKeywords(query)
	if err != nil {
		return nil, err
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	return memRepo.ReadByKeywordsAndImportance(uid, keywords, 0.5, limit)
}

func (*MemoryService) LoadRelevantMemories(uid, userMessage string, limit int) ([]*repository.Memory, error) {
	ms := &MemoryService{}
	keywords, err := ms.ExtractKeywords(userMessage)
	if err != nil {
		return ms.LoadMemories(uid, limit)
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	memories, err := memRepo.ReadByKeywordsAndImportance(uid, keywords, 0.5, limit)
	if err != nil {
		return nil, err
	}

	if len(memories) < limit/2 {
		generalMemories, err := memRepo.ReadByImportance(uid, 0.7, limit-len(memories))
		if err == nil {
			memories = append(memories, generalMemories...)
		}
	}

	return memories, nil
}
