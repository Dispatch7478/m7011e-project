package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Constants for RabbitMQ
const (
	ExchangeName = "t-hub.events"
	ExchangeType = "topic"
)

// Service holds the connection and channel
type Service struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Connect initializes the RabbitMQ connection and declares the exchange
func Connect() (*Service, error) {
	// 1. Get credentials and host from environment variables
	user := os.Getenv("RABBITMQ_DEFAULT_USER")
	pass := os.Getenv("RABBITMQ_DEFAULT_PASS")
	host := os.Getenv("RABBITMQ_HOST")

	if user == "" || pass == "" || host == "" {
		return nil, fmt.Errorf("RABBITMQ_DEFAULT_USER, RABBITMQ_DEFAULT_PASS, and RABBITMQ_HOST env vars must be set")
	}

	// 2. Construct the URL
	url := fmt.Sprintf("amqp://%s:%s@%s:5672/", user, pass, host)

	// 3. Dial the connection
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 4. Open a channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// 5. Declare the topic exchange
	err = ch.ExchangeDeclare(
		ExchangeName, // name
		ExchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	log.Println("Connected to RabbitMQ and exchange declared")
	return &Service{Conn: conn, Channel: ch}, nil
}

// Publish sends a message to the configured exchange with a routing key

func (s *Service) Publish(routingKey string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Channel.PublishWithContext(ctx,
		ExchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})

	if err != nil {
		return err
	}

	log.Printf("[x] Sent to %s: %s", routingKey, body)

	return nil
}
