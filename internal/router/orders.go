package router

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

func (r *Router) loadOrderHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Content-Type") != "text/plain" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		body := buf.String()

		order := utils.ParseOrder(body)
		if order == -1 {
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		err = r.svc.LoadOrder(req.Context(), order)
		if err != nil {
			switch err.Error() {
			case "200":
				res.WriteHeader(http.StatusOK)
				return
			case "409":
				res.WriteHeader(http.StatusConflict)
				return
			default:
				res.WriteHeader(http.StatusInternalServerError)
			}

		}

		res.WriteHeader(http.StatusAccepted)
	}
}

func (r *Router) getOrdersHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		orders, err := r.svc.GetOrders(req.Context())
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(*orders) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		resp, err := json.Marshal(orders)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Add("content-type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}
}
