package auth

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/zhenyangze/go-admin-build/goadmin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// AdministratorRole grants full access.
	AdministratorRole = "administrator"
)

// User is the built-in admin account model.
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"size:120;uniqueIndex"`
	Password  string `gorm:"size:255"`
	Name      string `gorm:"size:120"`
	Avatar    string `gorm:"size:255"`
	Roles     []Role `gorm:"many2many:admin_user_roles"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Role is the built-in RBAC role model.
type Role struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"size:120"`
	Slug        string       `gorm:"size:120;uniqueIndex"`
	Permissions []Permission `gorm:"many2many:admin_role_permissions"`
	Menus       []Menu       `gorm:"many2many:admin_role_menus"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission is the built-in permission model.
type Permission struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:120"`
	Slug      string `gorm:"size:120;uniqueIndex"`
	Path      string `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Menu is the built-in sidebar model.
type Menu struct {
	ID             uint `gorm:"primaryKey"`
	ParentID       uint
	Order          int
	Title          string `gorm:"size:120"`
	Icon           string `gorm:"size:120"`
	URI            string `gorm:"size:255"`
	PermissionSlug string `gorm:"size:120"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Service provides login, menu loading, and authorization over the built-in schema.
type Service struct {
	DB *gorm.DB
}

// NewService constructs the built-in auth service.
func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// AutoMigrate creates the built-in auth tables.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Role{}, &Permission{}, &Menu{})
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Authenticate validates a username/password pair.
func (s *Service) Authenticate(ctx context.Context, username, password string) (*goadmin.Identity, error) {
	var user User
	if err := s.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Roles.Menus").
		Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}
	return toIdentity(&user), nil
}

// FindIdentity reloads the current user from storage.
func (s *Service) FindIdentity(ctx context.Context, id uint) (*goadmin.Identity, error) {
	var user User
	if err := s.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Roles.Menus").
		First(&user, id).Error; err != nil {
		return nil, err
	}
	return toIdentity(&user), nil
}

// Navigation loads the menu tree for the current user.
func (s *Service) Navigation(ctx context.Context, identity *goadmin.Identity) ([]goadmin.NavigationItem, error) {
	var menus []Menu
	if hasRole(identity, AdministratorRole) {
		if err := s.DB.WithContext(ctx).Order("parent_id asc, `order` asc, id asc").Find(&menus).Error; err != nil {
			return nil, err
		}
	} else {
		var user User
		if err := s.DB.WithContext(ctx).
			Preload("Roles.Menus").
			First(&user, identity.ID).Error; err != nil {
			return nil, err
		}

		seen := map[uint]Menu{}
		for _, role := range user.Roles {
			for _, menu := range role.Menus {
				if menu.PermissionSlug != "" && !identity.Permissions[menu.PermissionSlug] {
					continue
				}
				seen[menu.ID] = menu
			}
		}
		for _, menu := range seen {
			menus = append(menus, menu)
		}
		sort.Slice(menus, func(i, j int) bool {
			if menus[i].ParentID == menus[j].ParentID {
				if menus[i].Order == menus[j].Order {
					return menus[i].ID < menus[j].ID
				}
				return menus[i].Order < menus[j].Order
			}
			return menus[i].ParentID < menus[j].ParentID
		})
	}

	itemsByParent := map[uint][]goadmin.NavigationItem{}
	for _, menu := range menus {
		itemsByParent[menu.ParentID] = append(itemsByParent[menu.ParentID], goadmin.NavigationItem{
			ID:       menu.ID,
			ParentID: menu.ParentID,
			Title:    menu.Title,
			Icon:     menu.Icon,
			URI:      menu.URI,
		})
	}
	return buildNavTree(itemsByParent, 0), nil
}

// Authorize checks whether a permission slug is granted.
func (s *Service) Authorize(_ context.Context, identity *goadmin.Identity, permission string) (bool, error) {
	if permission == "" {
		return true, nil
	}
	if identity == nil {
		return false, nil
	}
	if hasRole(identity, AdministratorRole) {
		return true, nil
	}
	return identity.Permissions[permission], nil
}

func toIdentity(user *User) *goadmin.Identity {
	identity := &goadmin.Identity{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.Name,
		Permissions: map[string]bool{},
	}
	for _, role := range user.Roles {
		identity.Roles = append(identity.Roles, role.Slug)
		for _, permission := range role.Permissions {
			identity.Permissions[permission.Slug] = true
		}
	}
	return identity
}

func hasRole(identity *goadmin.Identity, slug string) bool {
	for _, role := range identity.Roles {
		if role == slug {
			return true
		}
	}
	return false
}

func buildNavTree(itemsByParent map[uint][]goadmin.NavigationItem, parentID uint) []goadmin.NavigationItem {
	items := itemsByParent[parentID]
	for i := range items {
		items[i].Children = buildNavTree(itemsByParent, items[i].ID)
	}
	return items
}
