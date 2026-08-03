package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// otpTTL é o prazo para o visitante confirmar o código enviado por e-mail.
// Passado esse prazo, o cadastro pendente é considerado encerrado e é
// removido pela varredura periódica (ver DeleteExpiredPendingSignups).
const otpTTL = 2 * time.Hour

// otpResendCooldown evita reenvios em sequência (ex.: clique duplo).
const otpResendCooldown = 60 * time.Second

// otpMaxAttempts limita tentativas de código incorreto por cadastro pendente,
// para dificultar força bruta sobre o código de 6 dígitos.
const otpMaxAttempts = 5

var (
	// ErrPendingNotFound é retornado quando não há cadastro pendente para o e-mail.
	ErrPendingNotFound = errors.New("cadastro pendente não encontrado")
	// ErrOTPExpired é retornado quando o cadastro pendente já passou do prazo de confirmação.
	ErrOTPExpired = errors.New("código expirado")
	// ErrOTPInvalid é retornado quando o código informado não confere.
	ErrOTPInvalid = errors.New("código inválido")
	// ErrOTPLocked é retornado após exceder o número de tentativas permitido.
	ErrOTPLocked = errors.New("número de tentativas excedido, solicite um novo código")
	// ErrResendTooSoon é retornado ao tentar reenviar o código antes do intervalo mínimo.
	ErrResendTooSoon = errors.New("aguarde um pouco antes de solicitar um novo código")
)

// PendingSignup representa um cadastro aguardando confirmação do código OTP
// enviado por e-mail. A conta em users só é criada após a confirmação.
type PendingSignup struct {
	Email     string
	OTPSentAt time.Time
	ExpiresAt time.Time
}

// generateOTP gera um código numérico de 6 dígitos criptograficamente aleatório.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("gerando código OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// CreatePendingSignup grava (ou substitui, se já existir uma tentativa
// anterior não confirmada para o mesmo e-mail) um cadastro pendente e retorna
// o código OTP gerado, para ser enviado por e-mail pelo chamador.
func (r *Repository) CreatePendingSignup(ctx context.Context, name, email, passwordHash, role, cnpj string) (string, error) {
	code, err := generateOTP()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(otpTTL)

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO pending_signups (name, email, password_hash, role, cnpj, otp_code, otp_attempts, otp_sent_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			name = excluded.name,
			password_hash = excluded.password_hash,
			role = excluded.role,
			cnpj = excluded.cnpj,
			otp_code = excluded.otp_code,
			otp_attempts = 0,
			otp_sent_at = excluded.otp_sent_at,
			expires_at = excluded.expires_at
	`,
		name, email, passwordHash, role, nullableString(cnpj), code,
		now.Format(sqliteTimeFormat), expiresAt.Format(sqliteTimeFormat),
	)
	if err != nil {
		return "", fmt.Errorf("gravando cadastro pendente: %w", err)
	}
	return code, nil
}

// ResendPendingOTP gera um novo código para um cadastro pendente já
// existente, respeitando um intervalo mínimo entre reenvios. Não estende o
// prazo original de confirmação (expires_at permanece o mesmo).
func (r *Repository) ResendPendingOTP(ctx context.Context, email string) (string, error) {
	var sentAt, expiresAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT otp_sent_at, expires_at FROM pending_signups WHERE email = ?`, email,
	).Scan(&sentAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPendingNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lendo cadastro pendente: %w", err)
	}

	now := time.Now().UTC()
	if now.After(parseTimestamp(expiresAt)) {
		return "", ErrOTPExpired
	}
	if now.Sub(parseTimestamp(sentAt)) < otpResendCooldown {
		return "", ErrResendTooSoon
	}

	code, err := generateOTP()
	if err != nil {
		return "", err
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE pending_signups SET otp_code = ?, otp_attempts = 0, otp_sent_at = ? WHERE email = ?`,
		code, now.Format(sqliteTimeFormat), email,
	); err != nil {
		return "", fmt.Errorf("atualizando código OTP: %w", err)
	}
	return code, nil
}

// VerifyPendingSignup confirma o código informado e, se válido, cria a conta
// definitiva em users a partir dos dados do cadastro pendente, removendo-o em
// seguida. É a única forma de uma conta ser efetivamente criada.
func (r *Repository) VerifyPendingSignup(ctx context.Context, email, code string) (*User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("iniciando transação de confirmação: %w", err)
	}
	defer tx.Rollback()

	var name, passwordHash, role, storedCode string
	var cnpj sql.NullString
	var attempts int
	var expiresAt string
	err = tx.QueryRowContext(ctx,
		`SELECT name, password_hash, role, cnpj, otp_code, otp_attempts, expires_at FROM pending_signups WHERE email = ?`, email,
	).Scan(&name, &passwordHash, &role, &cnpj, &storedCode, &attempts, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPendingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lendo cadastro pendente: %w", err)
	}

	if time.Now().UTC().After(parseTimestamp(expiresAt)) {
		return nil, ErrOTPExpired
	}
	if attempts >= otpMaxAttempts {
		return nil, ErrOTPLocked
	}

	if code != storedCode {
		if _, err := tx.ExecContext(ctx, `UPDATE pending_signups SET otp_attempts = otp_attempts + 1 WHERE email = ?`, email); err != nil {
			return nil, fmt.Errorf("registrando tentativa inválida de OTP: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("confirmando tentativa inválida de OTP: %w", err)
		}
		if attempts+1 >= otpMaxAttempts {
			return nil, ErrOTPLocked
		}
		return nil, ErrOTPInvalid
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash, role, cnpj) VALUES (?, ?, ?, ?, ?)`,
		name, email, passwordHash, resolveSignupRole(email, role), cnpj,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("criando usuário confirmado: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do usuário confirmado: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM pending_signups WHERE email = ?`, email); err != nil {
		return nil, fmt.Errorf("removendo cadastro pendente confirmado: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("confirmando criação do usuário: %w", err)
	}

	return r.GetByID(ctx, id)
}

// DeleteExpiredPendingSignups encerra (remove) cadastros pendentes cujo prazo
// de confirmação já passou, chamada pela varredura periódica. Retorna quantos
// registros foram removidos.
func (r *Repository) DeleteExpiredPendingSignups(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM pending_signups WHERE expires_at < ?`, time.Now().UTC().Format(sqliteTimeFormat),
	)
	if err != nil {
		return 0, fmt.Errorf("removendo cadastros pendentes expirados: %w", err)
	}
	return res.RowsAffected()
}
