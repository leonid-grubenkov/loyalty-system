package models

type Withdraw struct {
	Number      string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt CustomTime `json:"processed_at"`
}
