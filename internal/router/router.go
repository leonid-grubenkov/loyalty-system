package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leonid-grubenkov/loyalty-system/internal/logging"
	"github.com/leonid-grubenkov/loyalty-system/internal/service"
)

type Router struct {
	Router *chi.Mux
	logger *logging.Logger
	svc    *service.Service
}

func NewRouter(logger *logging.Logger, svc *service.Service) *Router {
	r := Router{logger: logger, svc: svc}
	r.Router = chi.NewRouter()

	r.Router.Use(AuthHandle, logger.LoggingHandle)
	r.Router.Get("/ping", r.pingDBHandler())
	r.Router.Post("/api/user/register", r.registerHandler())
	r.Router.Post("/api/user/login", r.loginHandler())
	r.Router.Post("/api/user/orders", r.loadOrderHandler())
	r.Router.Get("/api/user/orders", r.getOrdersHandler())

	return &r
}

func (r *Router) pingDBHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := r.svc.Ping()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Header().Set("Content-Type", "text/plain")
			res.Write([]byte(err.Error()))
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Postgres connected!"))
	}
}
