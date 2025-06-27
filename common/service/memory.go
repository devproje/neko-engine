package service

import (
	"fmt"
	"os"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/util"
)

type MemoryService struct{}

type MemoryData struct {
	UID       string                `json:"user_id"`
	Histories []*repository.History `json:"histories"`
	// TODO: create memories variable here
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

	if err := db.GetDB().AutoMigrate(&repository.History{}); err != nil {
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
