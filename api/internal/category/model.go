package category

// Category representa uma categoria de transação, com padrões opcionais de
// auto-categorização (uma expressão regular por linha).
type Category struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
	Patterns string `json:"patterns"`
}
