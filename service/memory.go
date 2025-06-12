package service

import (
	"fmt"
	"os"
	"time"

	"github.com/devproje/neko-engine/model"
	"github.com/devproje/neko-engine/utils"
)

type Memory struct {
	Id         int
	UserID     string
	MemKey     string
	Content    string
	Importance int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MemoryService struct {
	maxCount       int
	expireDuration time.Duration
	minImportance  int

	Account *AccountService
	Gemini  *GeminiService
}

func init() {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer conn.Close()

	if _, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS memory (
			id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(25) NOT NULL,
			mem_key VARCHAR(40) NOT NULL,
			content TEXT NOT NULL,
			importance INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_tag (user_id, mem_key),
			FOREIGN KEY (user_id) REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE
		);
	`); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

func NewMemoryService(acc *AccountService, gemini *GeminiService) *MemoryService {
	return &MemoryService{
		maxCount:       10,
		expireDuration: 30 * 24 * time.Hour,
		minImportance:  5,
		Account:        acc,
		Gemini:         gemini,
	}
}

func (ms *MemoryService) create(memory *Memory) error {
	if memory.MemKey == "" || memory.Content == "" {
		_, _ = fmt.Fprintf(os.Stderr, "memory content is not contained values.\n")
		return nil
	}

	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("INSERT INTO memory (user_id, mem_key, content, importance) VALUES (?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(memory.UserID, memory.MemKey, memory.Content, memory.Importance); err != nil {
		return err
	}

	return ms.enforceMemoryLimit(memory.UserID)
}

func (ms *MemoryService) update(memory *Memory) error {
	if memory.MemKey == "" || memory.Content == "" {
		_, _ = fmt.Fprintf(os.Stderr, "memory content is not contained values.\n")
		return nil
	}

	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("UPDATE memory SET content = ?, importance = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND mem_key = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(memory.Content, memory.Importance, memory.UserID, memory.MemKey)
	return err
}

func (ms *MemoryService) enforceMemoryLimit(userID string) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT id FROM memory WHERE user_id = ? ORDER BY importance ASC, updated_at ASC")
	if err != nil {
		return err
	}

	rows, err := stmt.Query(userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	exceed := len(ids) - ms.maxCount
	if exceed <= 0 {
		return nil
	}

	_ = stmt.Close()

	stmt, err = conn.Prepare("DELETE FROM memory WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range exceed {
		if _, err := stmt.Exec(ids[i]); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MemoryService) Read(acc *model.Account) ([]*Memory, error) {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT id, user_id, mem_key, content, importance, created_at, updated_at FROM memory WHERE user_id = ? ORDER BY updated_at;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(acc.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var memory Memory
		if err := rows.Scan(&memory.Id, &memory.UserID, &memory.MemKey, &memory.Content, &memory.Importance, &memory.CreatedAt, &memory.UpdatedAt); err != nil {
			return nil, err
		}
		memories = append(memories, &memory)
	}
	return memories, nil
}

func (ms *MemoryService) SaveOrUpdate(acc *model.Account, memory *Memory) (string, error) {
	if memory == nil {
		return "", nil
	}

	memories, err := ms.Read(acc)
	if err != nil {
		return "", err
	}

	for _, exist := range memories {
		if exist.MemKey == memory.MemKey {
			if exist.Content == memory.Content && exist.Importance == memory.Importance {
				return "", nil
			}
			memory.Id = exist.Id
			if err = ms.update(memory); err != nil {
				return "", err
			}
			return "\n`✅ 기억 갱신완료!`", ms.enforceMemoryLimit(acc.Id)
		}
	}

	if err = ms.create(memory); err != nil {
		return "", err
	}
	return "\n`✅ 기억 갱신완료!`", ms.enforceMemoryLimit(acc.Id)
}
