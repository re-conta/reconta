package user

import (
	"slices"
	"time"
)

const (
	RolePessoaFisica   = "pessoa_fisica"
	RolePessoaJuridica = "pessoa_juridica"
	RoleContador       = "contador"
	RoleAdmin          = "admin"
	RoleSuperAdmin     = "super_admin"
)

// AssignableRoles são as roles que podem ser atribuídas manualmente por quem
// gerencia usuários. A role super_admin nunca é atribuível: é reservada aos
// e-mails de Super Admin.
var AssignableRoles = []string{RolePessoaFisica, RolePessoaJuridica, RoleContador, RoleAdmin}

// SignupRoles são as roles que um visitante pode escolher no cadastro.
var SignupRoles = []string{RolePessoaFisica, RolePessoaJuridica, RoleContador}

// Permissões administrativas atribuíveis por role. O Super Admin sempre tem
// todas, independentemente do que estiver gravado no banco.
const (
	PermAdminPanel        = "admin_panel"
	PermManageUsers       = "manage_users"
	PermManagePermissions = "manage_permissions"
)

var AllPermissions = []string{PermAdminPanel, PermManageUsers, PermManagePermissions}

// User representa um usuário cadastrado.
type User struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CNPJ        string    `json:"cnpj"`
	AvatarURL   string    `json:"avatarUrl"`
	HasPassword bool      `json:"hasPassword"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
}

// IsAdmin indica se o usuário tem acesso de administração (admin ou super_admin).
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// HasPermission indica se o usuário possui a permissão informada. O Super
// Admin sempre possui todas as permissões.
func (u *User) HasPermission(perm string) bool {
	if u.Role == RoleSuperAdmin {
		return true
	}
	return slices.Contains(u.Permissions, perm)
}
