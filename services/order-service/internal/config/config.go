package config

import (
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/database"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/client"
	"github.com/herodragmon/scalable-ecommerce/services/order-service/internal/rabbitmq"
)
type Config struct {
	DB            *database.Queries
	Platform      string
	ProductClient *client.ProductClient
	CartClient    *client.CartClient
	Publisher     *rabbitmq.Publisher
}
