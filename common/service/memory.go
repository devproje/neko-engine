package service

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/config"
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

	cnf := config.Load()
	model := "gemini-2.5-flash"
	if cnf != nil && cnf.Memory.Model != "" {
		model = cnf.Memory.Model
	}

	response, err := gemini.SendPrompt(systemPrompt, model, prompts)
	if err != nil {
		return nil, err
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	responseText := response.Candidates[0].Content.Parts[0].Text
	_, _ = fmt.Fprintf(os.Stderr, "Raw LLM response: %s\n", responseText)

	responseText = strings.TrimSpace(responseText)
	if after, found := strings.CutPrefix(responseText, "```json"); found {
		responseText = strings.TrimSuffix(after, "```")
		responseText = strings.TrimSpace(responseText)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Cleaned response for JSON parsing: %s\n", responseText)

	var analysis ImportanceAnalysis
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "JSON parsing failed, trying fallback analysis\n")
		ms := &MemoryService{}
		return ms.fallbackAnalysis(userMessage, botMessage, responseText)
	}

	if analysis.Importance < 0.0 {
		analysis.Importance = 0.0
	} else if analysis.Importance > 1.0 {
		analysis.Importance = 1.0
	}

	return &analysis, nil
}

func (ms *MemoryService) fallbackAnalysis(userMessage, botMessage, llmResponse string) (*ImportanceAnalysis, error) {
	_, _ = fmt.Fprintf(os.Stderr, "Using fallback analysis for: %s\n", llmResponse)

	keywords := ms.extractSimpleKeywords(userMessage + " " + botMessage)
	importance := ms.calculateBasicImportance(userMessage, botMessage)

	summary := fmt.Sprintf("User: %s... Bot: %s...",
		ms.truncateText(userMessage, 50),
		ms.truncateText(botMessage, 50))

	return &ImportanceAnalysis{
		Importance: importance,
		Summary:    summary,
		Reason:     "Fallback analysis due to LLM parsing failure",
		Keywords:   keywords,
	}, nil
}

func (ms *MemoryService) extractSimpleKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	keywordMap := make(map[string]bool)

	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(cleaned) >= 3 && !ms.isCommonWord(cleaned) {
			keywordMap[cleaned] = true
		}
	}

	keywords := make([]string, 0, len(keywordMap))
	for keyword := range keywordMap {
		keywords = append(keywords, keyword)
		if len(keywords) >= 7 { // 최대 7개
			break
		}
	}

	if len(keywords) < 3 {
		keywords = append(keywords, "conversation", "chat", "general")
	}

	return keywords
}

func (ms *MemoryService) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "you": true, "that": true, "was": true, "for": true,
		"are": true, "with": true, "his": true, "they": true, "this": true,
		"from": true, "not": true, "been": true, "have": true, "their": true, "said": true,
		"each": true, "which": true, "she": true, "how": true, "when": true, "can": true,
		"what": true, "where": true, "why": true, "who": true, "will": true, "more": true,
	}
	return commonWords[word]
}

func (ms *MemoryService) calculateBasicImportance(userMessage, botMessage string) float64 {
	score := 0.3

	if len(userMessage)+len(botMessage) > 200 {
		score += 0.2
	}

	combined := strings.ToLower(userMessage + " " + botMessage)
	importantWords := []string{"remember", "important", "prefer", "like", "hate", "love", "name", "age", "work", "hobby"}

	for _, word := range importantWords {
		if strings.Contains(combined, word) {
			score += 0.1
		}
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (ms *MemoryService) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func (ms *MemoryService) cleanText(text string) string {
	if !utf8.ValidString(text) {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid UTF-8 string detected, cleaning...\n")
		text = strings.ToValidUTF8(text, "?")
	}

	cleanRegex := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	text = cleanRegex.ReplaceAllString(text, "")

	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	text = strings.TrimSpace(text)

	if len(text) > 65535 {
		text = text[:65532] + "..."
	}

	return text
}

func (ms *MemoryService) cleanKeywords(keywords []string) []string {
	cleaned := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		cleanKeyword := ms.cleanText(keyword)
		if cleanKeyword != "" && len(cleanKeyword) <= 100 { // 키워드 길이 제한
			cleaned = append(cleaned, cleanKeyword)
		}
	}
	return cleaned
}

func (*MemoryService) SaveMemoryIfImportant(uid, userMessage, botMessage, providerID, providerUsername string) error {
	ms := &MemoryService{}
	analysis, err := ms.AnalyzeImportance(userMessage, botMessage)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to analyze importance: %v\n", err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stderr, "Memory analysis result - Importance: %.2f, Keywords: %v, Summary: %s\n",
		analysis.Importance, analysis.Keywords, analysis.Summary)

	cnf := config.Load()
	threshold := 0.5
	if cnf != nil {
		threshold = cnf.Memory.ImportanceThreshold
		if !cnf.Memory.Enable {
			_, _ = fmt.Fprintf(os.Stderr, "Memory saving disabled in config\n")
			return nil
		}
	}

	if analysis.Importance < threshold {
		_, _ = fmt.Fprintf(os.Stderr, "Memory not saved - importance %.2f < %.2f\n", analysis.Importance, threshold)
		return nil
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)

	cleanedUserMessage := ms.cleanText(userMessage)
	cleanedBotMessage := ms.cleanText(botMessage)
	cleanedSummary := ms.cleanText(analysis.Summary)
	cleanedKeywords := ms.cleanKeywords(analysis.Keywords)
	cleanedProviderUsername := ms.cleanText(providerUsername)

	keywordsStr := strings.Join(cleanedKeywords, ",")

	memory := &repository.Memory{
		UserID:           uid,
		UserMessage:      cleanedUserMessage,
		BotMessage:       cleanedBotMessage,
		Importance:       analysis.Importance,
		Summary:          cleanedSummary,
		Keywords:         keywordsStr,
		ProviderID:       providerID,
		ProviderUsername: cleanedProviderUsername,
	}

	_, _ = fmt.Fprintf(os.Stderr, "Saving memory with cleaned data - Summary length: %d, Keywords: %d\n",
		len(cleanedSummary), len(cleanedKeywords))

	return memRepo.Create(memory)
}

func (*MemoryService) UpdateMemory(memory *repository.Memory) error {
	ms := &MemoryService{}

	// 텍스트 데이터 정리
	memory.UserMessage = ms.cleanText(memory.UserMessage)
	memory.BotMessage = ms.cleanText(memory.BotMessage)
	memory.Summary = ms.cleanText(memory.Summary)
	memory.ProviderUsername = ms.cleanText(memory.ProviderUsername)

	// 키워드 정리
	if memory.Keywords != "" {
		keywords := strings.Split(memory.Keywords, ",")
		cleanedKeywords := ms.cleanKeywords(keywords)
		memory.Keywords = strings.Join(cleanedKeywords, ",")
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)

	_, _ = fmt.Fprintf(os.Stderr, "Updating memory with cleaned data - ID: %d, Summary length: %d\n",
		memory.ID, len(memory.Summary))

	return memRepo.Update(memory)
}

func (*MemoryService) ReanalyzeAndUpdateMemory(memoryID uint) error {
	ms := &MemoryService{}

	// 기존 메모리 조회
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	var memory repository.Memory
	if err := db.GetDB().First(&memory, memoryID).Error; err != nil {
		return fmt.Errorf("memory not found: %v", err)
	}

	// 중요도 재분석
	analysis, err := ms.AnalyzeImportance(memory.UserMessage, memory.BotMessage)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to reanalyze memory %d: %v\n", memoryID, err)
		return err
	}

	// 새로운 분석 결과로 업데이트
	memory.Importance = analysis.Importance
	memory.Summary = ms.cleanText(analysis.Summary)
	cleanedKeywords := ms.cleanKeywords(analysis.Keywords)
	memory.Keywords = strings.Join(cleanedKeywords, ",")

	_, _ = fmt.Fprintf(os.Stderr, "Reanalyzed memory %d - New importance: %.2f\n", memoryID, analysis.Importance)

	return memRepo.Update(&memory)
}

func (*MemoryService) DeleteMemory(memoryID uint) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	return memRepo.Delete(memoryID)
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

	cnf := config.Load()
	model := "gemini-2.5-flash"
	if cnf != nil && cnf.Memory.Model != "" {
		model = cnf.Memory.Model
	}

	response, err := gemini.SendPrompt(systemPrompt, model, prompts)
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

func (*MemoryService) LoadAllMemories(limit, offset int) ([]*repository.Memory, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var memories []*repository.Memory
	err := db.GetDB().Preload("User").Limit(limit).Offset(offset).Order("importance desc, created_at desc").Find(&memories).Error
	return memories, err
}

func (*MemoryService) GetMemoryByID(memoryID uint) (*repository.Memory, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var memory repository.Memory
	err := db.GetDB().Preload("User").First(&memory, memoryID).Error
	if err != nil {
		return nil, err
	}

	return &memory, nil
}

func (*MemoryService) FlushUserMemories(userID string) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	memRepo := repository.NewMemoryRepository(db)
	return memRepo.Flush(userID)
}

func (*MemoryService) GetMemoryStats() (map[string]interface{}, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var totalMemories int64
	var totalHistories int64
	var avgImportance float64
	var highImportanceCount int64

	db.GetDB().Model(&repository.Memory{}).Count(&totalMemories)
	db.GetDB().Model(&repository.History{}).Count(&totalHistories)
	db.GetDB().Model(&repository.Memory{}).Select("COALESCE(AVG(importance), 0)").Row().Scan(&avgImportance)
	db.GetDB().Model(&repository.Memory{}).Where("importance >= ?", 0.7).Count(&highImportanceCount)

	var topProviders []map[string]interface{}
	db.GetDB().Raw(`
		SELECT provider_username, COUNT(*) as memory_count
		FROM memories
		GROUP BY provider_username
		ORDER BY memory_count DESC
		LIMIT 10
	`).Scan(&topProviders)

	return map[string]interface{}{
		"total_memories":        totalMemories,
		"total_histories":       totalHistories,
		"average_importance":    avgImportance,
		"high_importance_count": highImportanceCount,
		"top_providers":         topProviders,
	}, nil
}
