package main

import (
	"net/http"

	"github.com/leonid-grubenkov/loyalty-system/internal/logging"
	"github.com/leonid-grubenkov/loyalty-system/internal/router"
	"github.com/leonid-grubenkov/loyalty-system/internal/service"
	"github.com/leonid-grubenkov/loyalty-system/internal/storage"
)

func main() {
	logger := logging.GetLogger()
	defer logger.Logger.Sync()

	db := storage.GetDB("postgres://loyalty:loyalty@localhost:5432/loyalty")
	defer db.DB.Close()

	svc := service.NewService(db)

	r := router.NewRouter(logger, svc)

	logger.Sl.Infow(
		"Starting server",
		"addr", ":8080",
	)
	err := http.ListenAndServe(":8080", r.Router)
	if err != nil {
		panic(err)
	}
}
