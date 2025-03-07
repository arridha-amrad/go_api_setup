package main

import (
	"log"
	"my-go-api/internal/config"
	"my-go-api/internal/routes"
	"my-go-api/internal/validation"
	"my-go-api/pkg/database"
	"my-go-api/pkg/utils"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	utils.SetTokenSecretKey(cfg.SecretKey)
	db, err := database.Connect(cfg.DB.DbUrl, cfg.DB.MaxIdleTime, cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("Database connection pool established")

	validate := validation.Init()
	router := routes.RegisterRoutes(db, validate)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
