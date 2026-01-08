package config

import "github.com/herodragmon/scalable-ecommerce/internal/database"

type APIConfig struct {
	DB        *database.Queries
	Platform  string
	JWTSecret string
}
