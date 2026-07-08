package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrEmailTaken é retornado ao tentar cadastrar um e-mail já existente.
var ErrEmailTaken = errors.New("e-mail já cadastrado")

// ErrNotFound é retornado quando o usuário não existe.
var ErrNotFound = errors.New("usuário não encontrado")

// ErrCannotModifyRole é retornado ao tentar alterar a role de um e-mail
// reservado a Super Admin, ou ao tentar atribuir a role super_admin manualmente.
var ErrCannotModifyRole = errors.New("não é possível alterar a role deste usuário")

// superAdminEmails são os e-mails que devem sempre ter a role super_admin,
// com poderes irrestritos e acesso ao painel de admin.
var superAdminEmails = map[string]bool{
	"sistematico@gmail.com": true,
	"lsbrum@icloud.com":     true,
}

// roleForEmail determina a role inicial de um usuário a partir do e-mail.
func roleForEmail(email string) string {
	if superAdminEmails[strings.ToLower(strings.TrimSpace(email))] {
		return RoleSuperAdmin
	}
	return RoleUser
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SyncSuperAdmins garante que os e-mails reservados de Super Admin estejam
// sempre com a role correta, mesmo que tenham sido criados antes desta
// funcionalidade existir ou alterados manualmente no banco.
func (r *Repository) SyncSuperAdmins(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = ? WHERE email IN (?, ?) AND role <> ?`,
		RoleSuperAdmin, "sistematico@gmail.com", "lsbrum@icloud.com", RoleSuperAdmin,
	)
	if err != nil {
		return fmt.Errorf("sincronizando super admins: %w", err)
	}
	return nil
}

// UpdateRole altera a role de um usuário para "user" ou "admin". A role
// super_admin nunca pode ser atribuída manualmente, e a role dos e-mails
// reservados de Super Admin nunca pode ser alterada.
func (r *Repository) UpdateRole(ctx context.Context, id int64, role string) (*User, error) {
	if role != RoleUser && role != RoleAdmin {
		return nil, ErrCannotModifyRole
	}

	u, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if superAdminEmails[strings.ToLower(u.Email)] {
		return nil, ErrCannotModifyRole
	}

	if _, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, id); err != nil {
		return nil, fmt.Errorf("atualizando role do usuário: %w", err)
	}
	return r.GetByID(ctx, id)
}

// UpdateProfile altera nome e e-mail do usuário.
func (r *Repository) UpdateProfile(ctx context.Context, id int64, name, email string) (*User, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET name = ?, email = ? WHERE id = ?`, name, email, id)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("atualizando perfil do usuário: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

// UpdatePassword troca a senha do usuário após validar a senha atual (bcrypt
// é comparado pelo chamador, que passa aqui apenas o novo hash). Usuários
// vinculados apenas ao Google (sem senha) definem a senha pela primeira vez
// através do mesmo fluxo, sem exigir senha atual — isso é decidido no handler.
func (r *Repository) UpdatePassword(ctx context.Context, id int64, newPasswordHash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, newPasswordHash, id)
	if err != nil {
		return fmt.Errorf("atualizando senha do usuário: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetPasswordHashByID retorna o hash de senha do usuário, usado para validar
// a senha atual antes de trocá-la. Retorna string vazia se o usuário não tem
// senha definida (cadastro apenas via Google).
func (r *Repository) GetPasswordHashByID(ctx context.Context, id int64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ?`, id).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lendo senha do usuário: %w", err)
	}
	return hash, nil
}

func (r *Repository) Create(ctx context.Context, name, email, passwordHash string) (*User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		name, email, passwordHash, roleForEmail(email),
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserindo usuário: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do usuário: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, role, avatar_url, password_hash <> '', created_at FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

// GetByEmailWithPasswordHash retorna o usuário e o hash de senha correspondente,
// usado exclusivamente pelo fluxo de autenticação.
func (r *Repository) GetByEmailWithPasswordHash(ctx context.Context, email string) (*User, string, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, role, avatar_url, created_at, password_hash FROM users WHERE email = ?`, email,
	)

	var u User
	var avatarURL sql.NullString
	var createdAt, passwordHash string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &avatarURL, &createdAt, &passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("lendo usuário: %w", err)
	}
	u.AvatarURL = avatarURL.String
	u.CreatedAt = parseTimestamp(createdAt)
	u.HasPassword = passwordHash != ""

	return &u, passwordHash, nil
}

// GetByGoogleID busca um usuário previamente vinculado a uma conta Google.
func (r *Repository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, role, avatar_url, password_hash <> '', created_at FROM users WHERE google_id = ?`, googleID,
	)
	return scanUser(row)
}

// GetByEmail busca um usuário pelo e-mail, usado para vincular uma conta Google
// a um cadastro já existente por e-mail/senha.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, role, avatar_url, password_hash <> '', created_at FROM users WHERE email = ?`, email,
	)
	return scanUser(row)
}

// CreateWithGoogle cria um usuário autenticado apenas via Google, sem senha.
func (r *Repository) CreateWithGoogle(ctx context.Context, name, email, googleID, avatarURL string) (*User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash, google_id, role, avatar_url) VALUES (?, ?, '', ?, ?, ?)`,
		name, email, googleID, roleForEmail(email), nullableString(avatarURL),
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserindo usuário via Google: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do usuário: %w", err)
	}

	return r.GetByID(ctx, id)
}

// LinkGoogleID vincula um google_id a um usuário existente (cadastrado antes via e-mail/senha).
func (r *Repository) LinkGoogleID(ctx context.Context, id int64, googleID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET google_id = ? WHERE id = ?`, googleID, id)
	if err != nil {
		return fmt.Errorf("vinculando conta Google: %w", err)
	}
	return nil
}

// UpdateAvatarURL atualiza a foto de perfil do usuário (ex.: avatar da conta
// Google), mantendo-a sincronizada a cada login.
func (r *Repository) UpdateAvatarURL(ctx context.Context, id int64, avatarURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET avatar_url = ? WHERE id = ?`, nullableString(avatarURL), id)
	if err != nil {
		return fmt.Errorf("atualizando avatar do usuário: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, role, avatar_url, password_hash <> '', created_at FROM users ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listando usuários: %w", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u, err := scanUserRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (*User, error) {
	var u User
	var avatarURL sql.NullString
	var createdAt string
	if err := s.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &avatarURL, &u.HasPassword, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo usuário: %w", err)
	}
	u.AvatarURL = avatarURL.String
	u.CreatedAt = parseTimestamp(createdAt)
	return &u, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanUserRows(rows *sql.Rows) (*User, error) {
	return scanUser(rows)
}

func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "constraint failed: UNIQUE")
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
