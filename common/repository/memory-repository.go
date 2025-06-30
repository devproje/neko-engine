package repository

import (
	"github.com/devproje/neko-engine/util"
	"gorm.io/gorm"
)

type Memory struct {
	UserID           string  `gorm:"index:idx_memory_user"`
	User             *User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserMessage      string  `gorm:"type:text"`
	BotMessage       string  `gorm:"type:text"`
	Importance       float64 `gorm:"type:decimal(3,2);index:idx_memory_importance"`
	Summary          string  `gorm:"type:text"`
	Keywords         string  `gorm:"type:text;index:idx_memory_keywords"`
	ProviderID       string  `gorm:"index:idx_memory_provider"`
	ProviderUsername string  `gorm:"index:idx_memory_provider_name"`
	gorm.Model
}

type MemoryRepository interface {
	Create(memory *Memory) error
	Read(uid string, limit int) ([]*Memory, error)
	ReadByImportance(uid string, minImportance float64, limit int) ([]*Memory, error)
	SearchByKeywords(uid string, keywords []string, limit int) ([]*Memory, error)
	ReadByKeywordsAndImportance(uid string, keywords []string, minImportance float64, limit int) ([]*Memory, error)
	Update(memory *Memory) error
	Delete(id uint) error
	Flush(uid string) error
}

type memoryRepository struct {
	db *util.Database
}

func NewMemoryRepository(database *util.Database) MemoryRepository {
	return &memoryRepository{db: database}
}

func (m *memoryRepository) Create(memory *Memory) error {
	return m.db.GetDB().Create(memory).Error
}

func (m *memoryRepository) Read(uid string, limit int) ([]*Memory, error) {
	var memories []*Memory
	err := m.db.GetDB().Where("user_id = ?", uid).Order("importance desc, created_at desc").Limit(limit).Find(&memories).Error
	return memories, err
}

func (m *memoryRepository) ReadByImportance(uid string, minImportance float64, limit int) ([]*Memory, error) {
	var memories []*Memory
	err := m.db.GetDB().Where("user_id = ? AND importance >= ?", uid, minImportance).Order("importance desc, created_at desc").Limit(limit).Find(&memories).Error
	return memories, err
}

func (m *memoryRepository) Update(memory *Memory) error {
	return m.db.GetDB().Save(memory).Error
}

func (m *memoryRepository) Delete(id uint) error {
	return m.db.GetDB().Delete(&Memory{}, id).Error
}

func (m *memoryRepository) SearchByKeywords(uid string, keywords []string, limit int) ([]*Memory, error) {
	if len(keywords) == 0 {
		return []*Memory{}, nil
	}
	
	var memories []*Memory
	query := m.db.GetDB().Where("user_id = ?", uid)
	
	for _, keyword := range keywords {
		query = query.Where("keywords LIKE ?", "%"+keyword+"%")
	}
	
	err := query.Order("importance desc, created_at desc").Limit(limit).Find(&memories).Error
	return memories, err
}

func (m *memoryRepository) ReadByKeywordsAndImportance(uid string, keywords []string, minImportance float64, limit int) ([]*Memory, error) {
	var memories []*Memory
	query := m.db.GetDB().Where("user_id = ? AND importance >= ?", uid, minImportance)
	
	if len(keywords) > 0 {
		for _, keyword := range keywords {
			query = query.Where("keywords LIKE ?", "%"+keyword+"%")
		}
	}
	
	err := query.Order("importance desc, created_at desc").Limit(limit).Find(&memories).Error
	return memories, err
}

func (m *memoryRepository) Flush(uid string) error {
	return m.db.GetDB().Where("user_id = ?", uid).Delete(&Memory{}).Error
}
