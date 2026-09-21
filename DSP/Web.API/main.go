package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"advertiser-api/api/controller"
	"advertiser-api/api/filter"
	"advertiser-api/model"
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
	var authorizationRepository repository.AuthorizationRepository
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
		authorizationRepository = repository.NewMySQLAuthorizationRepository(db)
	}

	authService := service.NewAuthService(loginRepository, sessionRepository)
	featureRepository := repository.NewMemoryFeatureRepository()
	for _, feature := range []struct{ code, name string }{
		{code: "offer", name: "广告内容与投放配置"},
		{code: "ads_partner", name: "广告商配置"},
		{code: "channel_partner", name: "渠道商配置"},
		{code: "channel_number", name: "渠道号配置"},
		{code: "channel_link", name: "渠道链接配置"},
		{code: "role", name: "角色管理"},
		{code: "feature", name: "功能权限管理"},
		{code: "report", name: "报表"},
		{code: "report.upload_log", name: "报表 / 上报日志"},
		{code: "report.settlement", name: "报表 / 结算报表"},
		{code: "report.attribution", name: "报表 / 归因报表"},
		{code: "account", name: "帐号管理"},
	} {
		for _, action := range []string{"create", "read", "update", "delete"} {
			if (feature.code == "report" || strings.HasPrefix(feature.code, "report.")) && action != "read" {
				continue
			}
			featureRepository.Create(model.CreateFeatureRequest{Code: feature.code + "." + action, Name: feature.name + " / " + action})
		}
	}
	roleRepository := repository.NewMemoryRoleRepository(featureRepository)
	if authorizationRepository == nil {
		authorizationRepository = repository.NewMemoryAuthorizationRepository(roleRepository)
	}
	authService.SetAuthorization(authorizationRepository)
	offerService := service.NewOfferService(offerRepository)
	accessService := service.NewAccessService(roleRepository, featureRepository)
	channelPartnerService := service.NewChannelPartnerService(repository.NewMemoryChannelPartnerRepository())
	adsPartnerService := service.NewAdsPartnerService(repository.NewMemoryAdsPartnerRepository())
	membershipController := controller.NewMembershipController(authService)
	offerController := controller.NewOfferController(offerService)
	accessController := controller.NewAccessController(accessService)
	channelPartnerController := controller.NewChannelPartnerController(channelPartnerService)
	adsPartnerController := controller.NewAdsPartnerController(adsPartnerService)
	healthController := controller.NewHealthController()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthController.Check)
	mux.HandleFunc("/api/v1/membership/login", membershipController.Login)
	mux.HandleFunc("/api/v1/membership/logout", membershipController.Logout)
	offerFilter := filter.Offer(authService, http.HandlerFunc(offerController.Collection))
	offerByIDFilter := filter.Offer(authService, http.HandlerFunc(offerController.ByID))
	mux.Handle("/api/v1/advertiser/offers", offerFilter)
	mux.Handle("/api/v1/advertiser/offers/", offerByIDFilter)
	roleFilter := func(next http.Handler) http.Handler { return filter.CRUD(authService, "role", next) }
	featureFilter := func(next http.Handler) http.Handler { return filter.CRUD(authService, "feature", next) }
	channelPartnerFilter := func(next http.Handler) http.Handler { return filter.CRUD(authService, "channel_partner", next) }
	adsPartnerFilter := func(next http.Handler) http.Handler { return filter.CRUD(authService, "ads_partner", next) }
	mux.Handle("/api/v1/roles", roleFilter(http.HandlerFunc(accessController.Roles)))
	mux.Handle("/api/v1/roles/", roleFilter(http.HandlerFunc(accessController.RoleByID)))
	mux.Handle("/api/v1/features", featureFilter(http.HandlerFunc(accessController.Features)))
	mux.Handle("/api/v1/features/", featureFilter(http.HandlerFunc(accessController.FeatureByID)))
	mux.Handle("/api/v1/channel-partners", channelPartnerFilter(http.HandlerFunc(channelPartnerController.Collection)))
	mux.Handle("/api/v1/channel-partners/options", channelPartnerFilter(http.HandlerFunc(channelPartnerController.Options)))
	mux.Handle("/api/v1/channel-partners/", channelPartnerFilter(http.HandlerFunc(channelPartnerController.ByID)))
	mux.Handle("/api/v1/ads-partners", adsPartnerFilter(http.HandlerFunc(adsPartnerController.Collection)))
	mux.Handle("/api/v1/ads-partners/", adsPartnerFilter(http.HandlerFunc(adsPartnerController.ByID)))
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
