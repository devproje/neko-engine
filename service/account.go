package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/model"
	"github.com/devproje/neko-engine/utils"
)

type AccountService struct{}

func init() {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer conn.Close()

	if _, err := conn.Exec(strings.TrimSpace(`CREATE TABLE IF NOT EXISTS accounts(
		id 			VARCHAR(25) NOT NULL,
		author 		VARCHAR(25) NOT NULL,
		role   		INT 		NOT NULL 	DEFAULT 1,
		prompt  	TEXT 		DEFAULT     '',
		count		INT 		NOT NULL	DEFAULT 	0,
		total		INT 		NOT NULL 	DEFAULT 	0,
		created_at 	DATETIME	DEFAULT		CURRENT_TIMESTAMP,
		updated_at	DATETIME    DEFAULT		CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		constraint	PK_Account	PRIMARY KEY (id)
	);`)); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

func NewAccountService() *AccountService {
	return &AccountService{}
}

func (*AccountService) Create(acc *model.Account) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("INSERT INTO accounts (id, author, role) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(acc.Id, acc.Author, acc.Role.Id); err != nil {
		return err
	}

	return nil
}

func (*AccountService) Read(id string) (*model.Account, error) {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT * FROM accounts WHERE id = ?;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user account id '%s' not found", id)
	}

	var roleId int
	var acc model.Account

	if err = rows.Scan(&acc.Id, &acc.Author, &roleId, &acc.Prompt, &acc.Count, &acc.Total, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
		return nil, err
	}

	role := model.RoleValueOf(roleId)
	acc.Role = *role

	return &acc, nil
}

func (*AccountService) UpdateName(acc *model.Account, name string) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("UPDATE accounts SET name = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(name, acc.Id); err != nil {
		return err
	}

	return nil
}

func (*AccountService) UpdatePersonalPrompt(acc *model.Account, prompt string) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("UPDATE accounts SET prompt = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(prompt, acc.Id); err != nil {
		return err
	}

	return nil
}

func (*AccountService) UpdateRole(acc *model.Account, roleId int) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("UPDATE accounts SET role = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(roleId, acc.Id); err != nil {
		return err
	}

	return nil
}

func (*AccountService) UpdateCount(acc *model.Account) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	var stmt *sql.Stmt
	stmt, err = conn.Prepare("UPDATE accounts SET count = count + 1, total = total + 1 WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(acc.Id); err != nil {
		return err
	}

	return nil
}

func (*AccountService) ResetCount() {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer conn.Close()

	if _, err = conn.Exec("update accounts set count = 0;"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

func (*AccountService) Delete() {}

func (*AccountService) Query() ([]model.Account, error) {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("select * from accounts;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var acc model.Account
		var roleId int
		if err = rows.Scan(&acc.Id, &acc.Author, &roleId, &acc.Prompt, &acc.Count, &acc.Total, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			return nil, err
		}

		role := model.RoleValueOf(roleId)
		acc.Role = *role
		accounts = append(accounts, acc)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("user accounts not found")
	}

	return accounts, nil
}

func (*AccountService) IsExist() bool {
	return true
}

func (*AccountService) ExchangeOAuth2CodeForToken(code string, tp int) (string, error) {
	cnf := config.Load()
	client := http.Client{}

	body := url.Values{}
	body.Set("code", code)
	body.Set("scope", "identify guilds")
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", cnf.Bot.ClientId)
	body.Set("client_secret", cnf.Bot.ClientSecret)
	body.Set("redirect_uri", fmt.Sprintf("%s?type=%d", cnf.Bot.RedirectURI, tp))

	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", bytes.NewBufferString(body.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s %d\n", string(buf), resp.StatusCode)
		return "", fmt.Errorf("failed to exchange code: status %d", resp.StatusCode)
	}

	var respData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}
	if respData.AccessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}
	return respData.AccessToken, nil
}

func (*AccountService) FetchUserInfoFromToken(token string) (map[string]any, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: status %d", resp.StatusCode)
	}

	var userInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	reqGuilds, err := http.NewRequest("GET", "https://discord.com/api/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	reqGuilds.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	respGuilds, err := http.DefaultClient.Do(reqGuilds)
	if err != nil {
		return nil, err
	}
	defer respGuilds.Body.Close()

	if respGuilds.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch guilds: status %d", respGuilds.StatusCode)
	}

	var guilds []interface{}
	if err := json.NewDecoder(respGuilds.Body).Decode(&guilds); err != nil {
		return nil, err
	}
	userInfo["guilds"] = guilds

	return userInfo, nil
}

func (*AccountService) FetchServerInfoFromToken(token string) (map[string]any, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_, _ = fmt.Fprintf(os.Stderr, "Error response body: %s\n", string(body))
		return nil, fmt.Errorf("failed to fetch server info: status %d", resp.StatusCode)
	}

	var guilds []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
		return nil, err
	}

	return map[string]any{"guilds": guilds}, nil
}

func (as *AccountService) HandleUserOAuth2Callback(code string) (bool, error) {
	cnf := config.Load()
	token, err := as.ExchangeOAuth2CodeForToken(code, 1)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	userInfo, err := as.FetchUserInfoFromToken(token)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	var roleId = 1
	guilds, ok := userInfo["guilds"].([]any)
	if !ok {
		return false, fmt.Errorf("failed to assert guilds from user info")
	}

	for _, g := range guilds {
		guild, ok := g.(map[string]any)
		if ok && fmt.Sprintf("%v", guild["id"]) == cnf.Bot.OfficialServerId {
			roleId = 2
			break
		}
	}

	acc := &model.Account{
		Id:     fmt.Sprintf("%v", userInfo["id"]),
		Author: fmt.Sprintf("%v", userInfo["name"]),
		Role:   *model.RoleValueOf(roleId),
	}

	if err := as.Create(acc); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return false, nil
	}

	return true, nil
}

func (as *AccountService) HandleServerOAuth2Callback(code string) error {
	token, err := as.ExchangeOAuth2CodeForToken(code, 0)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, err = as.FetchServerInfoFromToken(token)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
