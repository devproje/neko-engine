package repository

import (
	"slices"

	"github.com/devproje/neko-engine/util"
	"gorm.io/gorm"
)

type History struct {
	UserID  string `gorm:"index:idx_history_index"`
	User    *User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Content string
	Answer  string
	gorm.Model
}

type HistoryRepository interface {
	Create(history *History) error
	Read(uid string, limit int) ([]*History, error)
	PurgeOne(uid string) error
	PurgeN(uid string, n int) error
	Flush(uid string) error
}

type historyRepository struct {
	db *util.Database
}

func NewHistoryRepository(database *util.Database) HistoryRepository {
	return &historyRepository{db: database}
}

func (repo *historyRepository) Create(history *History) error {
	return repo.db.GetDB().Create(history).Error
}

func (repo *historyRepository) Read(uid string, limit int) ([]*History, error) {
	var list = make([]*History, 0)
	err := repo.db.GetDB().Where("user_id = ?", uid).Order("created_at desc").Limit(limit).Find(&list).Error

	slices.Reverse(list)
	return list, err
}

func (repo *historyRepository) PurgeOne(uid string) error {
	var history History
	err := repo.db.GetDB().Where("user_id = ?", uid).Order("created_at desc").Limit(1).First(&history).Error
	if err != nil {
		return err
	}

	return repo.db.GetDB().Delete(&history).Error
}

func (repo *historyRepository) PurgeN(uid string, n int) error {
	var histories []*History
	err := repo.db.GetDB().Where("user_id = ?", uid).Order("created_at desc").Limit(n).Find(&histories).Error
	if err != nil {
		return err
	}
	if len(histories) == 0 {
		return nil
	}
	ids := make([]uint, len(histories))
	for i, h := range histories {
		ids[i] = h.ID
	}

	return repo.db.GetDB().Delete(&History{}, ids).Error
}

func (repo *historyRepository) Flush(uid string) error {
	return repo.db.GetDB().Where("user_id = ?", uid).Delete(&History{}).Error
}
