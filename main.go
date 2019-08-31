package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Push(parent context.Context, key []byte, value []byte, w *kafka.Writer) (err error) {
	message := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	return w.WriteMessages(parent, message)
}

var w *kafka.Writer

func main() {

	kafkaconn, _ := kafka.DialLeader(context.Background(), "udp", "localhost:9092", "test", 5)
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test",
		Balancer: &kafka.LeastBytes{},
	})
	conn, err := amqp.Dial("amqp://guest:localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"q_name", // name
		true,                             // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,          // queue
		"test-consumer", // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range msgs {
			go func() {
				log.Printf("Recei§ved a message: %s", d.Body)
				w.WriteMessages(context.Background(),kafka.Message{Value: []byte(d.Body)})
			}()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	kafkaconn.Close()
}
