package models

type BalanceInfo struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type ctxKey string

const LoginKey ctxKey = "login"
