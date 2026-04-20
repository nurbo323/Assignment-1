package domain

type PaymentStats struct {
	TotalCount      int64 `json:"total_count"`
	AuthorizedCount int64 `json:"authorized_count"`
	DeclinedCount   int64 `json:"declined_count"`
	TotalAmount     int64 `json:"total_amount"`
}
