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
		{Id: 1, Name: "root", Limit: -1, Root: true},
		{Id: 2, Name: "user", Limit: 80, Root: false},
		{Id: 3, Name: "server", Limit: 120, Root: false},
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
	usr.Total++

	if err := user.Update(usr); err != nil {
		return err
	}

	return nil
}

func (as *AccountService) ResetCount() error {
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

func (*AccountService) IsRoot(userID string) (bool, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return false, err
	}
	defer db.Close()

	var user repository.User
	err := db.GetDB().Preload("Role").First(&user, "id = ?", userID).Error
	if err != nil {
		return false, err
	}

	if user.Role == nil {
		return false, fmt.Errorf("user has no role assigned")
	}

	return user.Role.Root, nil
}

func (*AccountService) IsUserRoot(user *repository.User) bool {
	if user.Role == nil {
		return false
	}
	return user.Role.Root
}

func (*AccountService) BanUser(userID string, adminID string) error {
	as := &AccountService{}
	isRoot, err := as.IsRoot(adminID)
	if err != nil {
		return fmt.Errorf("failed to verify admin permissions: %v", err)
	}
	if !isRoot {
		return fmt.Errorf("insufficient permissions: only root users can ban others")
	}

	user, err := as.ReadUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	if as.IsUserRoot(user) {
		return fmt.Errorf("cannot ban root users")
	}

	if user.Banned {
		return fmt.Errorf("user is already banned")
	}

	user.Banned = true
	if err := as.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to ban user: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "User %s (%s) has been banned by admin %s\n", user.Username, userID, adminID)
	return nil
}

func (*AccountService) UnbanUser(userID string, adminID string) error {
	as := &AccountService{}
	isRoot, err := as.IsRoot(adminID)
	if err != nil {
		return fmt.Errorf("failed to verify admin permissions: %v", err)
	}
	if !isRoot {
		return fmt.Errorf("insufficient permissions: only root users can unban others")
	}

	user, err := as.ReadUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	if !user.Banned {
		return fmt.Errorf("user is not banned")
	}

	user.Banned = false
	if err := as.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to unban user: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "User %s (%s) has been unbanned by admin %s\n", user.Username, userID, adminID)
	return nil
}

func (*AccountService) PromoteToRoot(userID string, adminID string) error {
	as := &AccountService{}
	isRoot, err := as.IsRoot(adminID)
	if err != nil {
		return fmt.Errorf("failed to verify admin permissions: %v", err)
	}
	if !isRoot {
		return fmt.Errorf("insufficient permissions: only root users can promote others")
	}

	user, err := as.ReadUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	if as.IsUserRoot(user) {
		return fmt.Errorf("user is already a root user")
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	role := repository.NewRoleRepository(db)
	rootRole, err := role.Read(1)
	if err != nil {
		return fmt.Errorf("failed to get root role: %v", err)
	}

	user.RoleID = rootRole.Id
	user.Role = rootRole
	if err := as.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to promote user: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "User %s (%s) has been promoted to root by admin %s\n", user.Username, userID, adminID)
	return nil
}

func (*AccountService) DemoteFromRoot(userID string, adminID string) error {
	as := &AccountService{}
	isRoot, err := as.IsRoot(adminID)
	if err != nil {
		return fmt.Errorf("failed to verify admin permissions: %v", err)
	}
	if !isRoot {
		return fmt.Errorf("insufficient permissions: only root users can demote others")
	}

	if userID == adminID {
		return fmt.Errorf("cannot demote yourself")
	}

	user, err := as.ReadUser(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	if !as.IsUserRoot(user) {
		return fmt.Errorf("user is not a root user")
	}

	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	role := repository.NewRoleRepository(db)
	userRole, err := role.Read(2)
	if err != nil {
		return fmt.Errorf("failed to get user role: %v", err)
	}

	user.RoleID = userRole.Id
	user.Role = userRole
	if err := as.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to demote user: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "User %s (%s) has been demoted from root by admin %s\n", user.Username, userID, adminID)
	return nil
}

func (*AccountService) ListUsers(limit, offset int) ([]*repository.User, int64, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, 0, err
	}
	defer db.Close()

	var users []*repository.User
	var total int64

	db.GetDB().Model(&repository.User{}).Count(&total)
	err := db.GetDB().Preload("Role").Limit(limit).Offset(offset).Order("created_at desc").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (*AccountService) SearchUsers(query string, limit, offset int) ([]*repository.User, int64, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, 0, err
	}
	defer db.Close()

	var users []*repository.User
	var total int64

	searchQuery := "%" + query + "%"
	whereClause := "username LIKE ? OR id LIKE ?"

	db.GetDB().Model(&repository.User{}).Where(whereClause, searchQuery, searchQuery).Count(&total)
	err := db.GetDB().Preload("Role").Where(whereClause, searchQuery, searchQuery).
		Limit(limit).Offset(offset).Order("created_at desc").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (*AccountService) ListRoles() ([]*repository.Role, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var roles []*repository.Role
	err := db.GetDB().Order("id asc").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (*AccountService) CreateRole(role *repository.Role) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	roleRepo := repository.NewRoleRepository(db)
	return roleRepo.Create(role)
}

func (*AccountService) UpdateRole(role *repository.Role) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	roleRepo := repository.NewRoleRepository(db)
	return roleRepo.Update(role)
}

func (*AccountService) DeleteRole(roleID int) error {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	roleRepo := repository.NewRoleRepository(db)
	return roleRepo.Delete(roleID)
}

func (*AccountService) GetUserStats() (map[string]interface{}, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var totalUsers int64
	var bannedUsers int64
	var rootUsers int64
	var totalChats int64

	db.GetDB().Model(&repository.User{}).Count(&totalUsers)
	db.GetDB().Model(&repository.User{}).Where("banned = ?", true).Count(&bannedUsers)
	db.GetDB().Model(&repository.User{}).Joins("JOIN roles ON users.role_id = roles.id").Where("roles.root = ?", true).Count(&rootUsers)
	db.GetDB().Model(&repository.User{}).Select("COALESCE(SUM(total), 0)").Row().Scan(&totalChats)

	return map[string]interface{}{
		"total_users":  totalUsers,
		"banned_users": bannedUsers,
		"root_users":   rootUsers,
		"total_chats":  totalChats,
	}, nil
}

func (*AccountService) GetRoleStats() (map[string]interface{}, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var roleStats []map[string]interface{}
	err := db.GetDB().Raw(`
		SELECT r.id, r.name, r.limit, r.root, COUNT(u.id) as user_count
		FROM roles r
		LEFT JOIN users u ON r.id = u.role_id
		GROUP BY r.id, r.name, r.limit, r.root
		ORDER BY r.id
	`).Scan(&roleStats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"role_distribution": roleStats,
	}, nil
}

func (*AccountService) SearchUsersByUsername(username string) ([]*repository.User, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var users []*repository.User
	searchQuery := "%" + username + "%"
	err := db.GetDB().Preload("Role").Where("username LIKE ?", searchQuery).Limit(10).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (*AccountService) GetTopUsersByTotal(limit int) ([]*repository.User, error) {
	db := util.NewDatabase()
	if err := db.Open(); err != nil {
		return nil, err
	}
	defer db.Close()

	var users []*repository.User
	err := db.GetDB().Preload("Role").Where("total > 0").Order("total desc").Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}
