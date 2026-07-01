package tag

// Tag representa uma etiqueta livre que pode ser associada a transações.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
