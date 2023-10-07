package models

import (
	"fmt"
	"time"
)

type Order struct {
	Number     string     `json:"number"`
	Login      string     `json:"login"`
	Status     string     `json:"status"`
	Accrual    int        `json:"accrual"`
	UploadedAt CustomTime `json:"uploaded_at"`
}

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	formatted := ct.Format(time.RFC3339)
	return []byte(`"` + formatted + `"`), nil
}

func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		ct.Time = t
		return nil
	}
	return fmt.Errorf("Unsupported type for CustomTime: %T", value)
}
