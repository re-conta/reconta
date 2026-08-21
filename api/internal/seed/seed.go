package seed

import (
	"context"
	"fmt"

	"github.com/re-conta/reconta/api/internal/account"
	"github.com/re-conta/reconta/api/internal/category"
)

type defaultCategory struct {
	name      string
	color     string
	icon      string
	typ       string
	isTaxable bool
}

// defaultCategories espelha as categorias criadas automaticamente na versão
// anterior (Next.js) do app para todo usuário novo. As receitas tipicamente
// tributáveis (salário, freelance) já entram marcadas para contar na base de
// cálculo do simulador de imposto de renda.
var defaultCategories = []defaultCategory{
	{"Alimentação", "#f97316", "utensils", "expense", false},
	{"Moradia", "#8b5cf6", "home", "expense", false},
	{"Transporte", "#3b82f6", "car", "expense", false},
	{"Saúde", "#ef4444", "heart", "expense", false},
	{"Educação", "#06b6d4", "book", "expense", false},
	{"Lazer", "#ec4899", "smile", "expense", false},
	{"Vestuário", "#a855f7", "shirt", "expense", false},
	{"Tecnologia", "#64748b", "laptop", "expense", false},
	{"Contas & Serviços", "#f59e0b", "zap", "expense", false},
	{"Outros Gastos", "#6b7280", "more-horizontal", "expense", false},
	{"Salário", "#10b981", "briefcase", "income", true},
	{"Freelance", "#14b8a6", "laptop", "income", true},
	{"Investimentos", "#22c55e", "trending-up", "income", false},
	{"Outros Ganhos", "#84cc16", "plus-circle", "income", false},
}

// Defaults cria as categorias e a conta padrão de um usuário recém-criado.
func Defaults(ctx context.Context, accounts *account.Repository, categories *category.Repository, userID int64) error {
	for _, c := range defaultCategories {
		if _, err := categories.Create(ctx, userID, c.name, c.color, c.icon, c.typ, "", c.isTaxable); err != nil {
			return fmt.Errorf("criando categoria padrão %q: %w", c.name, err)
		}
	}
	if _, err := accounts.Create(ctx, userID, "Conta Principal", "checking", 0); err != nil {
		return fmt.Errorf("criando conta padrão: %w", err)
	}
	return nil
}
