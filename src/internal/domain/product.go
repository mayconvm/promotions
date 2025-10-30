package domain

type Product struct {
	ProductID string `json:"id"`
	Title     string `json:"title"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
