package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "t-hub.events"
	ExchangeType = "topic"
)

type EventPublisher interface {
	Publish(routingKey string, body string) error
}

type Service struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func ConnectRabbitMQ() (*Service, error) {
	user := os.Getenv("RABBITMQ_DEFAULT_USER")
	pass := os.Getenv("RABBITMQ_DEFAULT_PASS")
	host := os.Getenv("RABBITMQ_HOST")

	url := fmt.Sprintf("amqp://%s:%s@%s:5672/", user, pass, host)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(ExchangeName, ExchangeType, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to RabbitMQ")
	return &Service{Conn: conn, Channel: ch}, nil
}

func (s *Service) Publish(routingKey string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.Channel.PublishWithContext(ctx, ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(body),
	})
}