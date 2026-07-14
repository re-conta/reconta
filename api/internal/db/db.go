package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open abre a conexão com o SQLite e garante que o schema exista.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("criando diretório do banco de dados: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrindo banco de dados: %w", err)
	}
	conn.SetMaxOpenConns(1) // SQLite: evita SQLITE_BUSY em escritas concorrentes

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("aplicando migrações: %w", err)
	}
	return conn, nil
}

func migrate(conn *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT NOT NULL,
		email         TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role          TEXT NOT NULL DEFAULT 'user',
		created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		token      TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS accounts (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name       TEXT NOT NULL,
		type       TEXT NOT NULL DEFAULT 'checking',
		balance    REAL NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS categories (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name     TEXT NOT NULL,
		color    TEXT NOT NULL DEFAULT '#6366f1',
		icon     TEXT NOT NULL DEFAULT 'circle',
		type     TEXT NOT NULL DEFAULT 'both',
		patterns TEXT
	);

	CREATE TABLE IF NOT EXISTS tags (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name    TEXT NOT NULL,
		color   TEXT NOT NULL DEFAULT '#6366f1'
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		date            TEXT NOT NULL,
		description     TEXT NOT NULL,
		amount          REAL NOT NULL,
		type            TEXT NOT NULL,
		category_id     INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		account_id      INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
		notes           TEXT,
		imported_from   TEXT,
		bank            TEXT,
		pix_beneficiary TEXT,
		created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS transaction_tags (
		transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
		tag_id         INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (transaction_id, tag_id)
	);

	CREATE TABLE IF NOT EXISTS monthly_opening_balances (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		month      INTEGER NOT NULL,
		year       INTEGER NOT NULL,
		amount     REAL NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		UNIQUE (user_id, month, year)
	);

	CREATE TABLE IF NOT EXISTS fixed_bills (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name        TEXT NOT NULL,
		amount      REAL NOT NULL,
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		account_id  INTEGER REFERENCES accounts(id) ON DELETE SET NULL,
		periodicity TEXT NOT NULL,
		due_date    TEXT NOT NULL,
		status      TEXT NOT NULL DEFAULT 'active',
		notes       TEXT,
		created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS fixed_bill_payments (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		fixed_bill_id  INTEGER NOT NULL REFERENCES fixed_bills(id) ON DELETE CASCADE,
		user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		due_date       TEXT NOT NULL,
		paid_at        TEXT NOT NULL,
		amount_paid    REAL NOT NULL,
		bank           TEXT,
		payment_method TEXT,
		notes          TEXT,
		transaction_id INTEGER REFERENCES transactions(id) ON DELETE SET NULL,
		created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		fixed_bill_id  INTEGER REFERENCES fixed_bills(id) ON DELETE CASCADE,
		kind           TEXT NOT NULL,
		title          TEXT NOT NULL,
		message        TEXT NOT NULL,
		due_date       TEXT NOT NULL,
		offset_minutes INTEGER NOT NULL,
		read_at        TEXT,
		email_sent_at  TEXT,
		created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		UNIQUE (fixed_bill_id, due_date, offset_minutes)
	);

	CREATE TABLE IF NOT EXISTS notification_settings (
		user_id       INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		site_enabled  INTEGER NOT NULL DEFAULT 1,
		email_enabled INTEGER NOT NULL DEFAULT 0,
		offsets       TEXT NOT NULL DEFAULT '[1440,120,60]',
		updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
	CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
	CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_user_id_date ON transactions(user_id, date);
	CREATE INDEX IF NOT EXISTS idx_fixed_bills_user_id ON fixed_bills(user_id);
	CREATE INDEX IF NOT EXISTS idx_fixed_bill_payments_fixed_bill_id ON fixed_bill_payments(fixed_bill_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id, read_at);
	`
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	return addMissingColumns(conn)
}

// addMissingColumns aplica alterações incrementais em tabelas já existentes,
// que CREATE TABLE IF NOT EXISTS não cobre (ex.: login via Google adicionado
// depois da tabela users já existir em bancos de produção).
func addMissingColumns(conn *sql.DB) error {
	hasColumn, err := columnExists(conn, "users", "google_id")
	if err != nil {
		return fmt.Errorf("verificando coluna google_id: %w", err)
	}
	if !hasColumn {
		if _, err := conn.Exec(`ALTER TABLE users ADD COLUMN google_id TEXT`); err != nil {
			return fmt.Errorf("adicionando coluna google_id: %w", err)
		}
		if _, err := conn.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)`); err != nil {
			return fmt.Errorf("criando índice único de google_id: %w", err)
		}
	}

	hasRole, err := columnExists(conn, "users", "role")
	if err != nil {
		return fmt.Errorf("verificando coluna role: %w", err)
	}
	if !hasRole {
		if _, err := conn.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`); err != nil {
			return fmt.Errorf("adicionando coluna role: %w", err)
		}
	}

	hasAvatarURL, err := columnExists(conn, "users", "avatar_url")
	if err != nil {
		return fmt.Errorf("verificando coluna avatar_url: %w", err)
	}
	if !hasAvatarURL {
		if _, err := conn.Exec(`ALTER TABLE users ADD COLUMN avatar_url TEXT`); err != nil {
			return fmt.Errorf("adicionando coluna avatar_url: %w", err)
		}
	}

	hasFixedBillPaymentID, err := columnExists(conn, "transactions", "fixed_bill_payment_id")
	if err != nil {
		return fmt.Errorf("verificando coluna fixed_bill_payment_id: %w", err)
	}
	if !hasFixedBillPaymentID {
		if _, err := conn.Exec(`ALTER TABLE transactions ADD COLUMN fixed_bill_payment_id INTEGER REFERENCES fixed_bill_payments(id) ON DELETE SET NULL`); err != nil {
			return fmt.Errorf("adicionando coluna fixed_bill_payment_id: %w", err)
		}
	}

	hasEmailSentAt, err := columnExists(conn, "notifications", "email_sent_at")
	if err != nil {
		return fmt.Errorf("verificando coluna email_sent_at: %w", err)
	}
	if !hasEmailSentAt {
		if _, err := conn.Exec(`ALTER TABLE notifications ADD COLUMN email_sent_at TEXT`); err != nil {
			return fmt.Errorf("adicionando coluna email_sent_at: %w", err)
		}
	}
	return nil
}

func columnExists(conn *sql.DB, table, column string) (bool, error) {
	rows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			colType    string
			notNull    int
			dfltValue  any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
