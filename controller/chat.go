package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/middleware"
	"github.com/devproje/neko-engine/model"
	"github.com/devproje/neko-engine/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ChatData struct {
	Id          string           `json:"id"`
	Author      string           `json:"author"`
	Content     string           `json:"content"`
	Attachments []ChatAttachment `json:"attachments"`
	Info        struct {
		GID  string `json:"gid"`
		CID  string `json:"cid"`
		NSFW bool   `json:"nsfw"`
	} `json:"info"`
}

type BotResponse struct {
	Answer           string `json:"answer"`
	MemoryTag        string `json:"memory_tag"`
	MemoryContent    string `json:"memory_content"`
	MemoryImportance int    `json:"memory_importance"`
	Sentiment        int    `json:"sentiment"`
}

type ChatAttachment struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
}

type ChatController struct {
	Gemini  *service.GeminiService
	Account *service.AccountService
	History *service.HistoryService
	Memory  *service.MemoryService
}

func NewChatController(
	gemini *service.GeminiService,
	account *service.AccountService,
	history *service.HistoryService,
	memory *service.MemoryService,
) *ChatController {
	return &ChatController{
		Gemini:  gemini,
		Account: account,
		History: history,
		Memory:  memory,
	}
}

func (cc *ChatController) getFileData(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	regex := regexp.MustCompile(`;\s*charset=[^;]*`)
	mimeType := regex.ReplaceAllString(resp.Header.Get("Content-Type"), "")
	if mimeType == "" {
		mimeType = "image/png"
	}
	return data, mimeType, nil
}

func (cc *ChatController) composePrompt(req *ChatData, acc *model.Account) (string, []*genai.Part, error) {
	var parts []*genai.Part

	id := fmt.Sprintf("<USER_PROFILE>\nUser's name is %s and ID is %s.\n</USER_PROFILE>\n\n", req.Author, req.Id)
	context := fmt.Sprintf(
		"<CURRENT_CONTEXT>\nCurrent time is %s and current channel is %s.\ncurrent affinity score is %d\n</CURRENT_CONTEXT>\n\n",
		time.Now().In(time.FixedZone("KST", 9*60*60)).Format("2006-01-02 15:04:05"),
		req.Info.CID,
		acc.Sentiment,
	)

	memories, _ := cc.Memory.Read(acc)
	mem := ""
	if len(memories) > 0 {
		mem += "<USER_MEMORY>\n"
		for _, m := range memories {
			mem += fmt.Sprintf("- [%s] %s (중요도: %d)\n", m.MemKey, m.Content, m.Importance)
		}
		mem += "</USER_MEMORY>\n\n"
	}

	hist, _ := cc.History.Load(acc, 15, req.Info.NSFW)
	history := "You must leverage all available conversation context in chronological order (from past to present),\n"
	history += "including previous dialogue and relevant metadata, to generate responses. \n"
	history += "Ensure your output demonstrates understanding of the ongoing user intent, prior exchanges, and the current situation.\n"
	history += "<HISTORY_METADATA>\n"
	for _, h := range hist {
		history += fmt.Sprintf("- [%s] user: %s\n- [%s] bot: %s\n",
			h.CreatedAt, h.User,
			h.CreatedAt, h.Bot,
		)
	}
	history += "</HISTORY_METADATA>\n"

	var prompt string
	pcnf := config.LoadPrompt()
	if pcnf == nil {
		return "", nil, fmt.Errorf("prompt config is not loaded")
	}

	prompt = pcnf.Default
	if pcnf.NSFW != "" && req.Info.NSFW {
		prompt = pcnf.NSFW
	}

	query := prompt + "\n" + id + context + mem + history
	parts = append(parts, genai.NewPartFromText(req.Content))

	for _, attr := range req.Attachments {
		data, mime, err := cc.getFileData(attr.URL)
		if err != nil {
			continue
		}
		parts = append(parts, genai.NewPartFromBytes(data, mime))
	}

	return query, parts, nil
}

func (cc *ChatController) Hit(ctx *gin.Context) {
	pcnf := config.LoadPrompt()
	if ok := middleware.CheckBot(ctx); !ok {
		return
	}

	start := time.Now()
	var req ChatData
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"errno": "invalid request"})
		return
	}

	acc, err := cc.Account.Read(req.Id)
	if err != nil {
		ctx.JSON(401, gin.H{"errno": "account is not registered"})
		return
	}

	if acc.Role.Id != model.RootRole.Id {
		if acc.Count > acc.Role.Limit {
			ctx.JSON(200, gin.H{
				"answer": "오늘의 사용한도를 초과 했어요. 내일 다시 시도 해주세요!",
			})
			return
		}
	}

	query, parts, err := cc.composePrompt(&req, acc)
	if err != nil {
		ctx.JSON(500, gin.H{"errno": "prompt composition failed"})
		return
	}

	prompts := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	resp, err := cc.Gemini.SendPrompt(query, pcnf.Model, prompts)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		ctx.JSON(500, gin.H{"errno": "Gemini response error"})
		return
	}

	result := resp.Text()
	exec := resp.ExecutableCode()
	codeResult := resp.CodeExecutionResult()

	if exec != "" {
		result += fmt.Sprintf("\n```\n%s\n```\n", exec)
		fmt.Println("[ChatController] executable code returned")

		if codeResult != "" {
			result += fmt.Sprintf("```bash\n%s\n```\n", codeResult)
			fmt.Println("[ChatController] code execution result returned")
		}
	}

	var indicator string
	compose := []*genai.Content{
		genai.NewContentFromText(fmt.Sprintf("user: %s\nbot: %s\n", req.Content, resp.Text()), genai.RoleUser),
	}

	memory, score, _ := cc.Gemini.AbstractDataFromLLM(acc, compose)
	if indicator, err = cc.Memory.SaveOrUpdate(acc, memory); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if err = cc.Account.UpdateSentiment(acc, score); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	cc.History.AppendLog(acc, req.Content, resp.Text(), req.Info.CID, req.Info.GID, req.Info.NSFW)
	fmt.Printf(
		"INPUT: %d, CANDIDATE: %d, TOTAL: %d\n",
		resp.UsageMetadata.PromptTokenCount, resp.UsageMetadata.CandidatesTokenCount, resp.UsageMetadata.TotalTokenCount,
	)

	if indicator != "" {
		indicator += fmt.Sprintf("\n`⏲️ 응답시간 (%.2fs), 호감도 반영: %d`", time.Since(start).Seconds(), score)
	} else {
		indicator = fmt.Sprintf("\n`⏲️ 응답시간 (%.2fs), 호감도 반영: %d`", time.Since(start).Seconds(), score)
	}

	ctx.JSON(200, gin.H{
		"answer": func() string {
			if !config.Debug {
				return result
			}

			return fmt.Sprintf("%s%s", result, indicator)
		}(),
		"usage": gin.H{
			"prompt":    resp.UsageMetadata.PromptTokenCount,
			"candidate": resp.UsageMetadata.CandidatesTokenCount,
			"total":     resp.UsageMetadata.TotalTokenCount,
		},
	})
}
