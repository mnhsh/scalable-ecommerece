package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/herodragmon/scalable-ecommerce/internal/admin"
	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/database"
	"github.com/herodragmon/scalable-ecommerce/internal/product"
	"github.com/herodragmon/scalable-ecommerce/internal/user"
)

func main() {
	godotenv.Load()

	platform := os.Getenv("PLATFORM")
	secretKey := os.Getenv("SECRET_KEY")
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}

	dbQueries := database.New(db)

	cfg := config.APIConfig{
		DB:        dbQueries,
		Platform:  platform,
		JWTSecret: secretKey,
	}

	const port = "8080"

	mux := http.NewServeMux()

	product.RegisterRoutes(mux, &cfg)
	user.RegisterRoutes(mux, &cfg)
	admin.RegisterRoutes(mux, &cfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(server.ListenAndServe())
}
