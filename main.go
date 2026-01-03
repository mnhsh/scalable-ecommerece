package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"

	"github.com/herodragmon/scalable-ecommerce/internal/database"
)


type apiConfig struct {
	db  *database.Queries
	platform string
	jwtSecret string
}

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

	cfg:= apiConfig{
		db: dbQueries,
		platform: platform,
		jwtSecret: secretKey,
	}

	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", cfg.handlerUsersCreate)

	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	server := &http.Server{
		Handler: mux,
		Addr: ":" + port,
	}
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(server.ListenAndServe())
}


