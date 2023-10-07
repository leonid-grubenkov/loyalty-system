package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/leonid-grubenkov/loyalty-system/internal/logging"
	"github.com/leonid-grubenkov/loyalty-system/internal/router"
	"github.com/leonid-grubenkov/loyalty-system/internal/service"
	"github.com/leonid-grubenkov/loyalty-system/internal/storage"
)

var options struct {
	flagRunAddr     string
	flagDatabaseDsn string
	flagAccrualAddr string
}

func main() {
	logger := logging.GetLogger()
	defer logger.Logger.Sync()

	parseFlags()

	db := storage.GetDB(options.flagDatabaseDsn)
	defer db.DB.Close()

	const numJobs = 10
	orders := make(chan int, numJobs)

	for w := 1; w <= 5; w++ {
		go service.Worker(w, options.flagAccrualAddr, orders)
	}

	svc := service.NewService(db, orders)

	r := router.NewRouter(logger, svc)

	logger.Sl.Infow(
		"Starting server",
		"addr", ":8080",
	)
	err := http.ListenAndServe(options.flagRunAddr, r.Router)
	if err != nil {
		panic(err)
	}
}

func parseFlags() {
	flag.StringVar(&options.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&options.flagAccrualAddr, "r", "http://localhost:8090", "address of accrual system")
	flag.StringVar(&options.flagDatabaseDsn, "d", "postgres://loyalty:loyalty@localhost:5432/loyalty", "database dsn")

	flag.Parse()
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		options.flagRunAddr = envRunAddr
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		options.flagAccrualAddr = envAccrualAddr
	}
	if envDatabaseDsn := os.Getenv("DATABASE_URI"); envDatabaseDsn != "" {
		options.flagDatabaseDsn = envDatabaseDsn
	}
}
