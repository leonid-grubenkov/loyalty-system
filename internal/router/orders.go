package router

import (
	"bytes"
	"log"
	"net/http"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

func (r *Router) loadOrderHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		user := req.Context().Value("login")
		log.Println(user)

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

		body := string(buf.Bytes())

		order := utils.ParseOrder(body)
		if order == -1 {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println(order)

		res.WriteHeader(http.StatusOK)
		return
	}
}
