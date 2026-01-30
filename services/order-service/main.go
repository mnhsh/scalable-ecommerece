package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/config"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/database"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/handlers"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/client"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/rabbitmq"
)

func main() {
	godotenv.Load()

	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
  if productServiceURL == "" {
		log.Fatal("PRODUCT_SERVICE_URL is not set")
  }
	productClient := client.NewProductClient(productServiceURL, 10*time.Second)
	
	cartServiceURL := os.Getenv("CART_SERVICE_URL")
	if cartServiceURL == "" {
		log.Fatal("CART_SERVICE_URL is not set")
	}
	cartClient := client.NewCartClient(cartServiceURL, 10*time.Second)

	dbQueries := database.New(db)

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Fatal("RABBITMQ_URL is not set")
	}

	publisher, err := rabbitmq.NewPublisher(rabbitmqURL)
	if err != nil {
		log.Fatalf("failed to publish exchange: %v", err)
	}
	defer publisher.Close()

	cfg := &config.Config{
		DB:       dbQueries,
		Platform: platform,
		ProductClient: productClient,
		CartClient: cartClient,
		Publisher: publisher,
	}

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, cfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Order service starting on port %s", port)
	log.Fatal(server.ListenAndServe())
}
