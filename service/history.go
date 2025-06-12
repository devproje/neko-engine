package service

import (
	"fmt"
	"os"
	"slices"

	"github.com/devproje/neko-engine/model"
	"github.com/devproje/neko-engine/utils"
)

type History struct {
	User      string
	Bot       string
	GID       string
	CID       string
	CreatedAt string
}

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

func init() {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer conn.Close()

	if _, err = conn.Exec(`CREATE TABLE IF NOT EXISTS chat_history(
		id        	BIGINT		PRIMARY KEY AUTO_INCREMENT,
		user_id   	VARCHAR(25)	NOT NULL,
		user   		TEXT		NOT NULL,
		bot   		TEXT		NOT NULL,
		nsfw	  	BOOL		NOT NULL DEFAULT FALSE,
		gid		  	VARCHAR(25)	NOT NULL,
		cid   	  	VARCHAR(25)	NOT NULL,
		created_at	DATETIME	NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX 		idx_user_time(user_id, created_at),
		CONSTRAINT FK_AccountHist_ID FOREIGN KEY(user_id) REFERENCES accounts(id)
			ON UPDATE CASCADE ON DELETE CASCADE
	);`); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

func (*HistoryService) AppendLog(info *model.Account, user, bot, cid, gid string, nsfw bool) error {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("INSERT INTO chat_history(user_id, user, bot, nsfw, gid, cid) VALUES (?, ?, ?, ?, ?, ?);")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(info.Id, user, bot, nsfw, gid, cid); err != nil {
		return err
	}

	return nil
}

func (s *HistoryService) Load(acc *model.Account, limit int, nsfw bool) ([]History, error) {
	db := utils.NewDatabase()
	conn, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	nfp := "and nsfw = 0"
	if nsfw {
		nfp = ""
	}

	stmt, err := conn.Prepare(fmt.Sprintf(`
		SELECT user, bot, gid, cid, created_at FROM chat_history 
			WHERE user_id = ? %s
			ORDER BY created_at DESC LIMIT ?;
	`, nfp))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(acc.Id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []History

	for rows.Next() {
		var user, bot, gid, cid, createdAt string
		if err := rows.Scan(&user, &bot, &gid, &cid, &createdAt); err != nil {
			return nil, err
		}

		history = append(history, History{
			User:      user,
			Bot:       bot,
			GID:       gid,
			CID:       cid,
			CreatedAt: createdAt,
		})
	}

	slices.Reverse(history)
	return history, nil
}
