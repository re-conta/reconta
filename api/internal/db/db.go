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
		created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
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

	CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
	CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
	CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_user_id_date ON transactions(user_id, date);
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
