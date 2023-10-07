package router

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/leonid-grubenkov/loyalty-system/internal/utils"
)

type LogPass struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (r *Router) registerHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var logPass LogPass
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &logPass); err != nil {
			log.Println(err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		err = r.svc.RegisterUser(logPass.Login, logPass.Password)
		if err != nil {
			if err.Error() == "user_already_exist" {
				r.logger.Sl.Infow("user_already_exist", "user: ", logPass.Login)
				res.WriteHeader(http.StatusConflict)
				return
			}
			r.logger.Sl.Errorf("register user err: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		jwt, err := utils.BuildJWTString(logPass.Login)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// expirationTime := time.Now().Add(60 * time.Minute)
		// http.SetCookie(res, &http.Cookie{
		// 	Name:    "token",
		// 	Value:   jwt,
		// 	Expires: expirationTime,
		// })
		res.Header().Set("Authorization", "Bearer "+jwt)

		res.WriteHeader(http.StatusOK)
		return
	}
}

func (r *Router) loginHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var logPass LogPass
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &logPass); err != nil {
			log.Println(err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		err = r.svc.LoginUser(logPass.Login, logPass.Password)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		jwt, err := utils.BuildJWTString(logPass.Login)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		// expirationTime := time.Now().Add(60 * time.Minute)
		// http.SetCookie(res, &http.Cookie{
		// 	Name:    "token",
		// 	Value:   jwt,
		// 	Expires: expirationTime,
		// })
		res.Header().Set("Authorization", "Bearer "+jwt)
		res.WriteHeader(http.StatusOK)
		return
	}

}
