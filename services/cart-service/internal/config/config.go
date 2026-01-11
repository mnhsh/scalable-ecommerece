package config

import (
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/database"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/client"
)
type Config struct {
	DB            *database.Queries
	Platform      string
	ProductClient *client.ProductClient
}
