package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/noggrj/autorepair/docs" // for swagger docs
	identityHttp "github.com/noggrj/autorepair/internal/identity/delivery/http"
	identityInfra "github.com/noggrj/autorepair/internal/identity/infrastructure"
	inventoryInfra "github.com/noggrj/autorepair/internal/inventory/infrastructure"
	notificationInfra "github.com/noggrj/autorepair/internal/notification/infrastructure"
	"github.com/noggrj/autorepair/internal/platform/config"
	"github.com/noggrj/autorepair/internal/platform/db"
	authMiddleware "github.com/noggrj/autorepair/internal/platform/middleware"
	serviceApp "github.com/noggrj/autorepair/internal/service/application"
	serviceHttp "github.com/noggrj/autorepair/internal/service/delivery/http"
	serviceInfra "github.com/noggrj/autorepair/internal/service/infrastructure"
)

// @version 1.0
// @description API for AutoRepair Shop Management System
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@autorepair.com

// @license.name Proprietary
// @license.url http://autorepair.com/license

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to DB
	database, err := db.New(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// 3. Setup Repositories
	userRepo := identityInfra.NewPostgresUserRepository(database.Pool)
	clientRepo := serviceInfra.NewPostgresClientRepository(database.Pool)
	vehicleRepo := serviceInfra.NewPostgresVehicleRepository(database.Pool)
	partRepo := inventoryInfra.NewPostgresPartRepository(database.Pool)
	serviceRepo := serviceInfra.NewPostgresServiceRepository(database.Pool)
	orderRepo := serviceInfra.NewPostgresOrderRepository(database.Pool)

	emailService := notificationInfra.NewConsoleEmailService()
	// ... other repos

	// 4. Setup Services
	orderService := serviceApp.NewOrderService(orderRepo, partRepo, clientRepo, emailService)

	// 5. Setup Handlers
	authHandler := identityHttp.NewAuthHandler(userRepo)
	clientHandler := serviceHttp.NewClientHandler(clientRepo)
	vehicleHandler := serviceHttp.NewVehicleHandler(vehicleRepo)
	partHandler := serviceHttp.NewPartHandler(partRepo)
	serviceHandler := serviceHttp.NewServiceHandler(serviceRepo)
	orderHandler := serviceHttp.NewOrderHandler(orderRepo, partRepo, serviceRepo, orderService)
	// ... other handlers

	// 6. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.Port)),
	))

	// Routes
	r.Group(func(r chi.Router) {
		// Public Routes
		r.Get("/orders/{id}/track", orderHandler.TrackOrder)
		r.Post("/orders/{id}/budget-response", orderHandler.ApproveBudget)

		r.Mount("/auth", func() http.Handler {
			sr := chi.NewRouter()
			authHandler.RegisterRoutes(sr)
			return sr
		}())

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AuthMiddleware)
			r.Mount("/admin", func() http.Handler {
				sr := chi.NewRouter()
				sr.Post("/clients", clientHandler.Create)
				sr.Get("/clients", clientHandler.List)
				sr.Put("/clients/{id}", clientHandler.Update)
				sr.Delete("/clients/{id}", clientHandler.Delete)

				sr.Post("/vehicles", vehicleHandler.Create)
				sr.Get("/vehicles", vehicleHandler.ListByClient)
				sr.Put("/vehicles/{id}", vehicleHandler.Update)
				sr.Delete("/vehicles/{id}", vehicleHandler.Delete)

				sr.Post("/parts", partHandler.Create)
				sr.Get("/parts", partHandler.List)

				sr.Post("/services", serviceHandler.Create)
				sr.Get("/services", serviceHandler.List)
				sr.Put("/services/{id}", serviceHandler.Update)
				sr.Delete("/services/{id}", serviceHandler.Delete)

				sr.Post("/orders", orderHandler.Create)
				sr.Get("/orders", orderHandler.ListActive)
				sr.Get("/orders/{id}", orderHandler.Get)
				sr.Patch("/orders/{id}/approve", orderHandler.Approve)
				sr.Post("/orders/{id}/diagnosis:start", orderHandler.StartDiagnosis)
				sr.Post("/orders/{id}/budget:send", orderHandler.SendBudget)
				sr.Post("/orders/{id}/finish", orderHandler.FinishOrder)
				sr.Post("/orders/{id}/deliver", orderHandler.DeliverOrder)
				sr.Patch("/orders/{id}/status", orderHandler.UpdateStatus)

				sr.Get("/reports/revenue", orderHandler.ReportRevenue)
				sr.Get("/reports/avg-execution-time", orderHandler.ReportAvgExecutionTime)

				return sr
			}())
		})
	})

	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
