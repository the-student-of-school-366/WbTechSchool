package main

import (
	"L3_1/internal/notify"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewProducer() (*Producer, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		failOnError(err, "Failed to make connection")
		return nil, err
	}
	chann, err := conn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return nil, err
	}
	err = chann.ExchangeDeclare(
		"notify", // name
		"fanout", // type
		false,    // durability
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		failOnError(err, "Failed to make an exchange")
		return nil, err
	}

	return &Producer{
		conn: conn,
		ch:   chann,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, notification notify.Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		failOnError(err, "Failed to marshal notification")
		return err
	}
	err = p.ch.PublishWithContext(ctx,
		"exchange", // exchange
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "applicatioin/json",
			Body:        body,
		})
	if err != nil {
		failOnError(err, "Failed to publish a message")
		return err
	}
	return nil
}

func (p *Producer) Close() error {
	err := p.ch.Close()
	if err != nil {
		return err
		failOnError(err, "Failed to close channel")
	}
	err = p.conn.Close()
	if err != nil {
		return err
		failOnError(err, "Failed to close connection")
	}
	return nil
}
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"notify", // name
		"fanout", // type
		false,    // durability
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := bodyFrom(os.Args)
	err = ch.PublishWithContext(ctx,
		"logs", // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "applicatioin/json",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
