package router

import (
	"encoding/json"
	"net/http"
)

type BalanceInfo struct {
	Balance   float64
	Withdrawn float64
}

func (r *Router) checkBalanceHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		info, err := r.svc.CheckBalance(req.Context())
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(info)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)

	}
}

func (r *Router) balanceWithdrawHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusOK)
	}
}

func (r *Router) checkWithdrawalsHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusOK)
	}
}
