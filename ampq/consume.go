package ampq

import (
	"github.com/streadway/amqp"
	"log"
)

func Consume(rabbit *Rabbit, callback func(msg string)) {
	log.Println("consume.rabbit_conf: ", rabbit)
	conn, err := amqp.Dial(rabbit.Url)
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}
	// declare exchange
	err = c.ExchangeDeclare(rabbit.Exchange, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("exchange.declare: %s", err)
	}
	// declare queue
	_, err = c.QueueDeclare(rabbit.Queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue.declare: %v", err)
	}

	// bind queue on exchange
	err = c.QueueBind(rabbit.Queue, rabbit.Queue, rabbit.Exchange, false, nil)
	if err != nil {
		log.Fatalf("queue.bind: %v", err)
	}

	// Set qos
	err = c.Qos(3, 0, false)
	if err != nil {
		log.Fatalf("basic.qos: %v", err)
	}

	// Consumer of tarantula
	deliveries, err := c.Consume(rabbit.Queue, "tarantula", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	// Services that consume messages
	// msg.Ack(false)
	go func() {
		for delivery := range deliveries {
			callback(string(delivery.Body))
			delivery.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	// blocking thread forever
	select {}
}
