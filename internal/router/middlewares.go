package router

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

const loginKey string = "login"

func AuthHandle(h http.Handler) http.Handler {

	jwtFn := func(w http.ResponseWriter, r *http.Request) {
		uri := strings.Split(r.RequestURI, "/")

		if (len(uri) == 2 && uri[1] == "ping") ||
			(len(uri) == 4 && (uri[3] == "login" || uri[3] == "register")) {

			h.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		tokenSplit := strings.Split(token, " ")
		if len(tokenSplit) != 2 || tokenSplit[0] != "Bearer" {
			http.Error(w, "empty token", http.StatusUnauthorized)
			return
		}

		login := utils.GetUserLogin(tokenSplit[1])
		if login == "" {
			log.Println("empty login")
			http.Error(w, "empty token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), loginKey, login)

		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(jwtFn)
}
