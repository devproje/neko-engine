package repository

import (
	"github.com/devproje/neko-engine/util"
	"gorm.io/gorm"
)

type Role struct {
	Id    int    `gorm:"primaryKey"`
	Name  string `gorm:"index"`
	Limit int
	Root  bool `gorm:"default:false"`
	gorm.Model
}

type RoleRepository interface {
	Create(role *Role) error
	Read(id int) (*Role, error)
	Update(role *Role) error
	Delete(id int) error
	Count() int
}

type roleRepository struct {
	db *util.Database
}

func NewRoleRepository(database *util.Database) RoleRepository {
	return &roleRepository{db: database}
}

func (repo *roleRepository) Create(role *Role) error {
	return repo.db.GetDB().Create(role).Error
}

func (repo *roleRepository) Read(id int) (*Role, error) {
	var role Role
	err := repo.db.GetDB().First(&role, id).Error

	return &role, err
}

func (repo *roleRepository) Update(role *Role) error {
	return repo.db.GetDB().Save(role).Error
}

func (repo *roleRepository) Delete(id int) error {
	return repo.db.GetDB().Delete(&Role{}, id).Error
}

func (repo *roleRepository) Count() int {
	var count int64
	repo.db.GetDB().Model(&Role{}).Count(&count)

	return int(count)
}
