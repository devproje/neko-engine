package service

import (
	"fmt"
	"os"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/util"
)

type AccountService struct{}

func NewAccountService() *AccountService {
	return &AccountService{}
}

func init() {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer db.Close()

	db.GetDB().AutoMigrate(&repository.Role{}, &repository.User{})
	role := repository.NewRoleRepository(db)
	role.Count()

	roles := []*repository.Role{
		{Id: 1, Name: "root", Limit: -1},
		{Id: 2, Name: "user", Limit: 80},
		{Id: 3, Name: "server", Limit: 120},
	}

	if role.Count() > 0 {
		return
	}

	for _, r := range roles {
		if err := role.Create(r); err == nil {
			continue
		}

		_, _ = fmt.Fprintf(os.Stderr, "\"%s\" role is already exist. ignore...\n", r.Name)
	}
}

func (*AccountService) CreateUser(id, author string) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	_ = db.GetDB().AutoMigrate(&repository.User{}, &repository.Role{})
	user := repository.NewUserRepository(db)
	role := repository.NewRoleRepository(db)
	def, _ := role.Read(2) // Get default role from database

	data := repository.User{
		ID:       id,
		Username: author,
		Role:     def,
	}

	if err := user.Create(&data); err != nil {
		return err
	}

	return nil
}

func (*AccountService) ReadUser(id string) (*repository.User, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	user := repository.NewUserRepository(db)
	ret, err := user.Read(id)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (*AccountService) UpdateUser(usr *repository.User) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	user := repository.NewUserRepository(db)
	if err := user.Update(usr); err != nil {
		return err
	}

	return nil
}

func (*AccountService) IncreaseCount(usr *repository.User) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	user := repository.NewUserRepository(db)
	usr.Count++

	if err := user.Update(usr); err != nil {
		return err
	}

	return nil
}

func (*AccountService) ResetCount() error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	user := repository.NewUserRepository(db)
	if err := user.ResetAll(); err != nil {
		return err
	}

	return nil
}

func (*AccountService) DeleteUser(id string) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	user := repository.NewUserRepository(db)
	if err := user.Delete(id); err != nil {
		return err
	}

	return nil
}

func (as *AccountService) GetRoleById(id int) (*repository.Role, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	role := repository.NewRoleRepository(db)

	return role.Read(id)
}
