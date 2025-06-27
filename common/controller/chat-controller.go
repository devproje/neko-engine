package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ChatController struct {
	Account *service.AccountService
	Gemini  *service.GeminiService
	Memory  *service.MemoryService
	Prompt  *service.PromptService
}

type ChatForm struct {
	Id          string       `json:"id"`
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

func NewChatController(
	account *service.AccountService,
	gemini *service.GeminiService,
	memory *service.MemoryService,
	prompt *service.PromptService,
) *ChatController {
	return &ChatController{Gemini: gemini, Prompt: prompt, Account: account}
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

func (cc *ChatController) composeSystemPrompt(acc *repository.User, role *repository.Role, persona *service.NKFile, req *ChatForm) string {
	var prompt string
	system := persona.Prompt.Default
	if req.Info.NSFW && persona.Prompt.NSFW != "" {
		system = persona.Prompt.NSFW
	}

	prompt += fmt.Sprintf("%s\n", system)
	prompt += fmt.Sprintf("<USER_PROFILE>\nCurrent user name is %s and ID is %s.</USER_PROFILE>\n\n", acc.Username, role.Name)
	prompt += fmt.Sprintf("<CURRENT_CONTEXT>\nCurrent timestamp is %d\n</CURRENT_CONTEXT>\n\n", time.Now().Unix())

	mem, err := cc.Memory.LoadHistory(acc.ID)
	if err != nil {
		return prompt
	}

	if len(mem.Histories) <= 0 {
		return prompt
	}

	prompt += "You must leverage all available conversation context in chronological order (from past to present),\n"
	prompt += "including previous dialogue and relevant metadata, to generate responses. \n"
	prompt += "Ensure your output demonstrates understanding of the ongoing user intent, prior exchanges, and the current situation.\n"
	prompt += "<HISTORY_METADATA>"
	for _, hist := range mem.Histories {
		prompt += fmt.Sprintf("- [%s] user: %s\n- [%s] bot: %s\n",
			hist.CreatedAt, hist.Content,
			hist.CreatedAt, hist.Answer,
		)
	}
	prompt += "</HISTORY_METADATA>"

	return prompt
}

func (cc *ChatController) SendChat(ctx *gin.Context) {
	var req ChatForm

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "some required parameter is not contained",
		})
		return
	}

	account, err := cc.Account.ReadUser(req.Id)
	if err != nil {
		ctx.JSON(401, gin.H{
			"errno": "Please sign up before using the bot!",
		})
		return
	}

	role, _ := cc.Account.GetRoleById(account.RoleID)
	if account.Count+1 > role.Limit {
		ctx.JSON(403, gin.H{
			"errno": "You have reached your chat limit for this role.",
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

	input := make([]*genai.Content, 0)
	parts := make([]*genai.Part, 0)

	if len(req.Attachments) != 0 {
		for _, attach := range req.Attachments {
			raw, mime, err := cc.getFileData(attach.URL)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}

			parts = append(parts, genai.NewPartFromBytes(raw, mime))
		}
	}

	input = append(input, genai.NewContentFromText(req.Content, genai.RoleUser))

	if len(req.Attachments) != 0 {
		input = append(input, genai.NewContentFromParts(parts, genai.RoleUser))
	}

	prompt := cc.composeSystemPrompt(account, role, persona, &req)
	resp, err := cc.Gemini.SendPrompt(prompt, persona.Model, input)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Gemini API is not responding",
		})
		return
	}

	var answer = resp.Text()
	cc.Memory.AppendHistory(&repository.History{
		UserID:  req.Id,
		Content: req.Content,
		Answer:  answer,
	})

	if err = cc.Account.IncreaseCount(account); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to increase user chat count",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"answer": answer,
		"usage": gin.H{
			"prompt":    resp.UsageMetadata.PromptTokenCount,
			"candidate": resp.UsageMetadata.CandidatesTokenCount,
			"total":     resp.UsageMetadata.TotalTokenCount,
		},
	})
}
