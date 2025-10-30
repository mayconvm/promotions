package domain

type Session struct {
	SessionId    string   `json:"id"`
	CronSchedule string   `json:"cron_schedule"`
	ProviderIds  []string `json:"provider_ids"`
	ProductIds   []string `json:"product_ids"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}
