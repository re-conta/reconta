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

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrindo banco de dados: %w", err)
	}
	// Com WAL, leituras concorrem entre si e com um escritor; escritas
	// simultâneas esperam via busy_timeout em vez de falhar com SQLITE_BUSY.
	conn.SetMaxOpenConns(8)

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

	CREATE TABLE IF NOT EXISTS role_permissions (
		role       TEXT NOT NULL,
		permission TEXT NOT NULL,
		PRIMARY KEY (role, permission)
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

	CREATE TABLE IF NOT EXISTS financial_health_settings (
		id                INTEGER PRIMARY KEY CHECK (id = 1),
		enabled           INTEGER NOT NULL DEFAULT 1,
		threshold_otima   REAL NOT NULL DEFAULT 20,
		threshold_boa     REAL NOT NULL DEFAULT 10,
		threshold_estavel REAL NOT NULL DEFAULT 0,
		threshold_ruim    REAL NOT NULL DEFAULT -10,
		updated_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS plans (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		code          TEXT NOT NULL UNIQUE,
		name          TEXT NOT NULL,
		description   TEXT NOT NULL DEFAULT '',
		price_monthly REAL NOT NULL DEFAULT 0,
		price_yearly  REAL NOT NULL DEFAULT 0,
		benefits      TEXT NOT NULL DEFAULT '[]',
		highlight     INTEGER NOT NULL DEFAULT 0,
		sort_order    INTEGER NOT NULL DEFAULT 0,
		updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id                   INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id              INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		plan_id              INTEGER NOT NULL REFERENCES plans(id),
		cycle                TEXT NOT NULL,
		status               TEXT NOT NULL DEFAULT 'pending',
		payment_method       TEXT NOT NULL,
		started_at           TEXT,
		current_period_end   TEXT,
		cancel_at_period_end INTEGER NOT NULL DEFAULT 0,
		canceled_at          TEXT,
		refund_amount        REAL,
		last_reminder_days   INTEGER,
		created_at           TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at           TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE TABLE IF NOT EXISTS subscription_payments (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
		user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		mp_payment_id   INTEGER,
		amount          REAL NOT NULL,
		method          TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'pending',
		status_detail   TEXT NOT NULL DEFAULT '',
		pix_qr          TEXT,
		pix_qr_base64   TEXT,
		ticket_url      TEXT,
		created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);

	CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions(user_id, status);
	CREATE INDEX IF NOT EXISTS idx_subscription_payments_mp ON subscription_payments(mp_payment_id);
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

	hasCNPJ, err := columnExists(conn, "users", "cnpj")
	if err != nil {
		return fmt.Errorf("verificando coluna cnpj: %w", err)
	}
	if !hasCNPJ {
		if _, err := conn.Exec(`ALTER TABLE users ADD COLUMN cnpj TEXT`); err != nil {
			return fmt.Errorf("adicionando coluna cnpj: %w", err)
		}
	}

	// A role legada "user" virou "pessoa_fisica" quando os cargos do site
	// foram introduzidos (Pessoa Física, Pessoa Jurídica, Contador, ...).
	if _, err := conn.Exec(`UPDATE users SET role = 'pessoa_fisica' WHERE role = 'user'`); err != nil {
		return fmt.Errorf("migrando role legada 'user': %w", err)
	}

	if err := seedDefaultRolePermissions(conn); err != nil {
		return err
	}
	return seedDefaultPlans(conn)
}

// seedDefaultRolePermissions grava as permissões padrão das roles uma única
// vez (controlado via PRAGMA user_version, para não sobrescrever edições
// feitas depois no painel de admin). Por padrão apenas a role admin tem
// acesso ao painel; o super_admin tem todas as permissões implicitamente.
func seedDefaultRolePermissions(conn *sql.DB) error {
	var version int
	if err := conn.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("lendo user_version: %w", err)
	}
	if version >= 1 {
		return nil
	}

	if _, err := conn.Exec(`
		INSERT OR IGNORE INTO role_permissions (role, permission) VALUES
			('admin', 'admin_panel'),
			('admin', 'manage_users')
	`); err != nil {
		return fmt.Errorf("populando permissões padrão: %w", err)
	}

	if _, err := conn.Exec(`PRAGMA user_version = 1`); err != nil {
		return fmt.Errorf("gravando user_version: %w", err)
	}
	return nil
}

// seedDefaultPlans grava os planos padrão do site (um gratuito e dois pagos)
// na primeira execução. INSERT OR IGNORE preserva preços e benefícios já
// editados no painel de admin em execuções seguintes.
func seedDefaultPlans(conn *sql.DB) error {
	if _, err := conn.Exec(`
		INSERT OR IGNORE INTO plans (code, name, description, price_monthly, price_yearly, benefits, highlight, sort_order) VALUES
			('gratuito', 'Gratuito', 'Para organizar as finanças do dia a dia', 0, 0,
			 '["Transações e categorias ilimitadas","1 conta bancária","Tags e filtros","Relatórios do mês atual"]', 0, 1),
			('essencial', 'Essencial', 'Para quem quer o controle completo', 19.90, 199.00,
			 '["Tudo do plano Gratuito","Contas bancárias ilimitadas","Importação de extratos (PDF/OFX)","Contas fixas com lembretes por e-mail","Relatórios completos e exportação"]', 1, 2),
			('profissional', 'Profissional', 'Para MEIs, empresas e contadores', 39.90, 399.00,
			 '["Tudo do plano Essencial","Saúde financeira detalhada","Relatórios em PDF, XLSX e ODS","Gestão de múltiplos CNPJs","Suporte prioritário"]', 0, 3)
	`); err != nil {
		return fmt.Errorf("populando planos padrão: %w", err)
	}

	// A permissão manage_plans foi introduzida junto com o sistema de planos;
	// user_version 2 concede ao admin uma única vez, sem sobrescrever edições.
	var version int
	if err := conn.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("lendo user_version: %w", err)
	}
	if version >= 2 {
		return nil
	}
	if _, err := conn.Exec(`INSERT OR IGNORE INTO role_permissions (role, permission) VALUES ('admin', 'manage_plans')`); err != nil {
		return fmt.Errorf("concedendo manage_plans ao admin: %w", err)
	}
	if _, err := conn.Exec(`PRAGMA user_version = 2`); err != nil {
		return fmt.Errorf("gravando user_version: %w", err)
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
