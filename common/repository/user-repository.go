package repository

import (
	"github.com/devproje/neko-engine/util"
	"gorm.io/gorm"
)

type User struct {
	ID        string `gorm:"primarykey"`
	Username  string `gorm:"index"`
	Prompt    string
	RoleID    int
	Role      *Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Count     int   `gorm:"default:0"`
	Total     int   `gorm:"default:0"`
	Sentiment int   `gorm:"default:0"`
	Banned    bool  `gorm:"default:false"`
	gorm.Model
}

type UserRepository interface {
	Create(usr *User) error
	Read(id string) (*User, error)
	Update(usr *User) error
	Delete(id string) error
	ResetAll() error
}

type userRepository struct {
	db *util.Database
}

func NewUserRepository(database *util.Database) UserRepository {
	return &userRepository{db: database}
}

func (repo *userRepository) Create(usr *User) error {
	return repo.db.GetDB().Create(usr).Error
}

func (repo *userRepository) Read(id string) (*User, error) {
	var user User
	err := repo.db.GetDB().First(&user, id).Error

	return &user, err
}

func (repo *userRepository) Update(usr *User) error {
	return repo.db.GetDB().Save(usr).Error
}

func (repo *userRepository) Delete(id string) error {
	return repo.db.GetDB().Delete(&User{}, id).Error
}

func (repo *userRepository) ResetAll() error {
	return repo.db.GetDB().Model(&User{}).Where("1 = 1").Update("count", 0).Error
}
