package user

import "time"

const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

// User representa um usuário cadastrado.
type User struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	AvatarURL   string    `json:"avatarUrl"`
	HasPassword bool      `json:"hasPassword"`
	CreatedAt   time.Time `json:"createdAt"`
}

// IsAdmin indica se o usuário tem acesso de administração (admin ou super_admin).
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}
