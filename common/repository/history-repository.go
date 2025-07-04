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
