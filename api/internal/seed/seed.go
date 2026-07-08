package seed

import (
	"context"
	"fmt"

	"github.com/re-conta/reconta/api/internal/account"
	"github.com/re-conta/reconta/api/internal/category"
)

type defaultCategory struct {
	name  string
	color string
	icon  string
	typ   string
}

// defaultCategories espelha as categorias criadas automaticamente na versão
// anterior (Next.js) do app para todo usuário novo.
var defaultCategories = []defaultCategory{
	{"Alimentação", "#f97316", "utensils", "expense"},
	{"Moradia", "#8b5cf6", "home", "expense"},
	{"Transporte", "#3b82f6", "car", "expense"},
	{"Saúde", "#ef4444", "heart", "expense"},
	{"Educação", "#06b6d4", "book", "expense"},
	{"Lazer", "#ec4899", "smile", "expense"},
	{"Vestuário", "#a855f7", "shirt", "expense"},
	{"Tecnologia", "#64748b", "laptop", "expense"},
	{"Contas & Serviços", "#f59e0b", "zap", "expense"},
	{"Outros Gastos", "#6b7280", "more-horizontal", "expense"},
	{"Salário", "#10b981", "briefcase", "income"},
	{"Freelance", "#14b8a6", "laptop", "income"},
	{"Investimentos", "#22c55e", "trending-up", "income"},
	{"Outros Ganhos", "#84cc16", "plus-circle", "income"},
}

// Defaults cria as categorias e a conta padrão de um usuário recém-criado.
func Defaults(ctx context.Context, accounts *account.Repository, categories *category.Repository, userID int64) error {
	for _, c := range defaultCategories {
		if _, err := categories.Create(ctx, userID, c.name, c.color, c.icon, c.typ, ""); err != nil {
			return fmt.Errorf("criando categoria padrão %q: %w", c.name, err)
		}
	}
	if _, err := accounts.Create(ctx, userID, "Conta Principal", "checking", 0); err != nil {
		return fmt.Errorf("criando conta padrão: %w", err)
	}
	return nil
}
