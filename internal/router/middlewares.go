package router

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

func AuthHandle(h http.Handler) http.Handler {

	jwtFn := func(w http.ResponseWriter, r *http.Request) {
		method := (strings.Split(r.RequestURI, "/"))[3]
		if method == "login" || method == "register" {
			h.ServeHTTP(w, r)
			return
		}
		token, err := r.Cookie("token")
		if err != nil {
			log.Println("empty token")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		login := utils.GetUserLogin(token.Value)
		if login == "" {
			log.Println("empty login")
			return
		}

		ctx := context.WithValue(r.Context(), "login", login)

		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(jwtFn)
}
