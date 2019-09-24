package main

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"log"
	"os"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func closeOnFinish(conn *amqp.Connection, ch *amqp.Channel, kafkaConn *kafka.Conn) {
	connErr := conn.Close()
	failOnError(connErr, "Closing connection is failed!")
	chErr := ch.Close()
	failOnError(chErr, "Closing channel is failed!")
	kafkaConnErr := kafkaConn.Close()
	failOnError(kafkaConnErr, "Closing kafka connection is failed!")
}

func main() {
	queueUrl := os.Getenv("RABBITMQ_URL")
	queuePort := os.Getenv("RABBITMQ_PORT")
	queueName := os.Getenv("QUEUE_NAME")
	kafkaconn, _ := kafka.DialLeader(context.Background(), "udp", "kafka:9092", "test", 1)

	fmt.Println("queueUrl: " + queueUrl + ":" + queuePort)
	fmt.Println("queueName: " + queueName)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9092"},
		Topic:    "test",
		Balancer: &kafka.LeastBytes{},
	})
	conn, err := amqp.Dial("amqp://" + queueUrl + ":" + queuePort + "/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer closeOnFinish(conn, ch, kafkaconn)

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,                // queue
		queueName+"-consumer", // consumer
		true,                  // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range msgs {
			go func() {
				log.Printf("Received a message: %s", d.Body)
				err := w.WriteMessages(context.Background(), kafka.Message{Value: []byte(d.Body)})
				failOnError(err, "Failed to write to kafka topic.")
			}()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
