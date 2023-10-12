package logging

import (
	"net/http"
	"time"

	"github.com/leonid-grubenkov/loyalty-system/internal/models"
	"go.uber.org/zap"
)

type Logger struct {
	Sl     *zap.SugaredLogger
	Logger *zap.Logger
}

func GetLogger() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sl := logger.Sugar()
	return &Logger{Sl: sl, Logger: logger}
}

func (l *Logger) LoggingHandle(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		user := r.Context().Value(models.LoginKey)

		l.Sl.Infoln(
			"URI", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
			"user", user,
		)
	}
	return http.HandlerFunc(logFn)
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.responseData.status = statusCode // захватываем код статуса
	r.ResponseWriter.WriteHeader(statusCode)
}
