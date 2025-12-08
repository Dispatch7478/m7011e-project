// Just learning how rabbitmq works.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Declare the queue name globally or passed as config
const QueueName = "test_queue"

// Service holds the connection and channel
type Service struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Connect initializes the RabbitMQ connection
func Connect() (*Service, error) {
	// 1. Get credentials and host from environment variables
	user := os.Getenv("RABBITMQ_DEFAULT_USER")
	pass := os.Getenv("RABBITMQ_DEFAULT_PASS")
	host := os.Getenv("RABBITMQ_HOST")

	if user == "" || pass == "" || host == "" {
		return nil, fmt.Errorf("RABBITMQ_DEFAULT_USER, RABBITMQ_DEFAULT_PASS, and RABBITMQ_HOST environment variables must be set")
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

	// 5. Declare the queue (idempotent - creates if not exists)
	_, err = ch.QueueDeclare(
		QueueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Println("Connected to RabbitMQ")
	return &Service{Conn: conn, Channel: ch}, nil
}

// Publish sends a message to the queue
func (s *Service) Publish(body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Channel.PublishWithContext(ctx,
		"",        // exchange
		QueueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	
	if err != nil {
		return err
	}
	log.Printf("[x] Sent: %s", body)
	return nil
}

// StartConsumer starts a goroutine to listen for messages
func (s *Service) StartConsumer() error {
	msgs, err := s.Channel.Consume(
		QueueName, // queue
		"",        // consumer tag (empty = auto-generated)
		true,      // auto-ack (true for simple testing)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	// Run in a goroutine so it doesn't block your main API
	go func() {
		log.Println("[*] Waiting for messages...")
		for d := range msgs {
			log.Printf("[!!] RECEIVED MESSAGE: %s", d.Body)
		}
	}()

	return nil
}