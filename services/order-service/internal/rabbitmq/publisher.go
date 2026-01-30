package rabbitmq

import (
	"fmt"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange string
}

func NewPublisher(url string) (*Publisher, error) {
	fmt.Println("order-service connecting to RabbitMQ")

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("could not connect to RabbitMQ: %w", err)
	}
	fmt.Println("order-service connected to RabbitMQ")

	publishCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("could not create channel: %w", err)
	}
	
	exchangeName := "orders"
	err = publishCh.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not declare exchange: %w", err)
	}
	return &Publisher{
		conn:     conn,
		channel:  publishCh,
		exchange: exchangeName,
	}, nil
}

func (p *Publisher) Publish(routingKey string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cant marshal the struct: %w", err)
	}

	err = p.channel.Publish(
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		return fmt.Errorf("cant publish order event: %w", err)
	}
	return nil
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}


