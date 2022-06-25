package ampq

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

// Publish is used to publish message to rabbitmq
func Publish(rabbit *Rabbit, message string) {
	log.Printf("publish.rabbit_conf: %v, \n publish.message: %v\n", rabbit, message)

	conn, err := amqp.Dial(rabbit.Url)
	if err != nil {
		log.Fatalf("connection.open: %s", err)
	}

	// This waits for a server acknowledgment which means the sockets will have
	// flushed all outbound publishings prior to returning.  It's important to
	// block on Close to not lose any publishings.
	defer conn.Close()

	c, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
	}

	err = c.ExchangeDeclare(rabbit.Exchange, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("exchange.declare: %v", err)
	}
	// declare queue and bind it to exchange
	_, err = c.QueueDeclare(rabbit.Queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue.declare: %v", err)
	}

	err = c.QueueBind(rabbit.Queue, rabbit.Queue, rabbit.Exchange, true, nil)

	if err != nil {
		log.Fatalf("queue.bind: %v", err)
	}

	// Prepare this message to be persistent.  Your publishing requirements may
	// be different.
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte(message),
	}

	// This is not a mandatory delivery, so it will be dropped if there are no
	// queues bound to the logs exchange.
	err = c.Publish(rabbit.Exchange, rabbit.Queue, false, false, msg)
	if err != nil {
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		log.Fatalf("basic.publish: %v", err)
	}
	fmt.Println("Publish finished!")
}
