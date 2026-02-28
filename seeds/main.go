package main

import (
	"log"

	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/identity/infrastructure"
	"github.com/noggrj/autorepair/internal/platform/config"
	"github.com/noggrj/autorepair/internal/platform/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := db.New(cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := infrastructure.NewPostgresUserRepository(pool.Pool)

	adminEmail := "admin@autorepair.com"
	adminPassword := "admin123"

	// Check if admin exists
	_, err = userRepo.GetByEmail(adminEmail)
	if err == nil {
		log.Println("Admin user already exists")
		return
	}

	user, err := domain.NewUser("Admin User", adminEmail, adminPassword, domain.RoleAdmin)
	if err != nil {
		log.Fatal(err)
	}

	if err := userRepo.Save(user); err != nil {
		log.Fatal(err)
	}

	log.Println("Admin user created successfully")
}
