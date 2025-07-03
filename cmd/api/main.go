package main

import (
	"log"
	"my-go-api/internal/config"
	"my-go-api/internal/routes"
	"my-go-api/internal/validation"
	"my-go-api/pkg/database"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	rdb := database.ConnectRedis(cfg.RDB.ADDR, cfg.RDB.Password, cfg.RDB.DB)
	if err != nil {
		log.Panic(err)
	}

	db, err := database.Connect(cfg.DB.DbUrl, cfg.DB.MaxIdleTime, cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns)
	if err != nil {
		log.Panic(err)
	}

	defer rdb.Close()
	defer db.Close()

	validate := validation.Init()
	router := routes.RegisterRoutes(db, rdb, validate, cfg)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
