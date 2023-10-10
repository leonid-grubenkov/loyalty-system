package router

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

type BalanceInfo struct {
	Balance   float64
	Withdrawn float64
}

type WithdrawnInfo struct {
	Order string  `josn:"order"`
	Sum   float64 `json:"sum"`
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

		var withdrawnInfo WithdrawnInfo
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &withdrawnInfo); err != nil {
			log.Println(err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		orderNum, err := utils.ParseOrder(withdrawnInfo.Order)
		if err != nil {
			log.Println("err when parse order num: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if orderNum == -1 {
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if withdrawnInfo.Sum == 0 {
			http.Error(res, "invalid sum", http.StatusBadRequest)
			return
		}

		err = r.svc.NewWithdrawn(req.Context(), orderNum, withdrawnInfo.Sum)
		if err != nil {
			switch err.Error() {
			case "insufficient balance":
				res.WriteHeader(http.StatusPaymentRequired)
				return
			case "already registred":
				res.WriteHeader(http.StatusConflict)
			default:
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusOK)
	}
}

func (r *Router) getWithdrawalsHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		withdrawns, err := r.svc.GetWithdrawns(req.Context())
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(*withdrawns) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		resp, err := json.Marshal(withdrawns)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Add("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}
}
