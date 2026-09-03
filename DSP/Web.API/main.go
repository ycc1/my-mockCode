package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"advertiser-api/api/controller"
	"advertiser-api/api/filter"
	"advertiser-api/repository"
	"advertiser-api/service"

	_ "github.com/go-sql-driver/mysql"
)

func routes() http.Handler {
	handler, err := buildRoutes()
	if err != nil {
		panic(err)
	}
	return handler
}

func buildRoutes() (http.Handler, error) {
	offerRepository := repository.NewMemoryOfferRepository()
	sessionRepository := repository.NewMemorySessionRepository()
	username := os.Getenv("API_USERNAME")
	password := os.Getenv("API_PASSWORD")
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin123"
	}
	var loginRepository repository.CredentialRepository = repository.NewMemoryCredentialRepository(username, password)
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("open mysql: %w", err)
		}
		pingContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(pingContext)
		cancel()
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("connect mysql: %w", err)
		}
		tableName := os.Getenv("MYSQL_LOGIN_TABLE")
		if tableName == "" {
			tableName = "users"
		}
		loginRepository, err = repository.NewMySQLLoginRepository(db, tableName)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	authService := service.NewAuthService(loginRepository, sessionRepository)
	offerService := service.NewOfferService(offerRepository)
	membershipController := controller.NewMembershipController(authService)
	offerController := controller.NewOfferController(offerService)
	healthController := controller.NewHealthController()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthController.Check)
	mux.HandleFunc("/api/v1/membership/login", membershipController.Login)
	mux.HandleFunc("/api/v1/membership/logout", membershipController.Logout)
	offerFilter := filter.Offer(authService, http.HandlerFunc(offerController.Collection))
	offerByIDFilter := filter.Offer(authService, http.HandlerFunc(offerController.ByID))
	mux.Handle("/api/v1/advertiser/offers", offerFilter)
	mux.Handle("/api/v1/advertiser/offers/", offerByIDFilter)
	return logging(mux), nil
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })
}

func main() {
	log.Println("advertiser API listening on :8080")
	handler, err := buildRoutes()
	if err != nil {
		log.Fatal(err)
	}
	if err := http.ListenAndServe(":8080", handler); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
