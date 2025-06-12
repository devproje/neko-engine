package model

type Account struct {
	Id        string `json:"id"`
	Author    string `json:"author"`
	Role      Role   `json:"role"`
	Prompt    string `json:"prompt"`
	Count     int    `json:"count"`
	Total     int    `json:"total"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Role struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Admin bool   `json:"admin"`
}

var (
	RootRole Role = Role{
		Id:    0,
		Name:  "root",
		Limit: -1,
		Admin: true,
	}
	UserRole Role = Role{
		Id:    1,
		Name:  "user",
		Limit: 35,
		Admin: false,
	}
	ServerRole Role = Role{
		Id:    2,
		Name:  "server",
		Limit: 60,
		Admin: false,
	}
)

func NewAccount(id, name string, server bool) *Account {
	return &Account{
		Id:     id,
		Author: name,
		Role: func() Role {
			if server {
				return ServerRole
			}

			return UserRole
		}(),
	}
}

func RoleValueOf(id int) *Role {
	switch id {
	case 0:
		return &RootRole
	case 1:
		return &UserRole
	case 2:
		return &ServerRole
	}

	return nil
}
