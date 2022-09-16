package middleware

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Connection This is a method encapsulation of RabbitMQ message publishing and message receiving,
// which supports Consumer disconnection and reconnection
// The current package depends on the {@link github.com/rabbitmq/amqp091-go} package
type Connection struct {
	Url string
	// The exchange name
	Exchange string
	// The exchange type
	ExchangeType string
	// The queue name
	Queue string
	// Whether to enable reconnection(only consumer)
	Retry bool
	// Retry times
	RetryTimes int
	// Retry interval
	RetryInterval time.Duration
	// The consumer callback function
	ConsumerCallback func(message string)
	// Timeout when publishing a message, default 5s
	PublishTimeout time.Duration
}

// Consumer This is a method used to start a rabbitMQ consumer
func (c *Connection) Consumer() {
	conn, err := amqp.Dial(c.Url)
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		FailOnError(err, "Failed to close a connection")
	}(conn)

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		FailOnError(err, "Failed to close a channel")
	}(ch)

	q, err := ch.QueueDeclare(
		c.Queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	FailOnError(err, "Failed to set QoS")

	if len(c.Exchange) > 0 {
		// exchange is empty， use the default exchange of rabbitMQ
		// else declare exchange and bind queue on it
		// declare queue and bind it to exchange
		err = ch.ExchangeDeclare(
			c.Exchange,
			c.ExchangeType,
			true,
			false,
			false,
			false,
			nil)
		FailOnError(err, "Failed to declare exchange")

		err = ch.QueueBind(q.Name, q.Name, c.Exchange, true, nil)
		FailOnError(err, "Failed to bind queue to exchange")
	}

	// Connection error notify
	conError := make(chan *amqp.Error)
	conn.NotifyClose(conError)

	// Channel error notify
	chError := make(chan *amqp.Error)
	ch.NotifyClose(chError)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	go func() {
		for {
			select {
			case d := <-msgs:
				if len(d.Body) > 0 {
					c.ConsumerCallback(string(d.Body))
					d.Ack(false)
				}
			case <-conError:
			case <-chError:
				if c.Retry {
					c.ConsumerReConnect(c.RetryTimes)
				}
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// ConsumerReConnect This a method is called to reconnect consumer
func (c *Connection) ConsumerReConnect(times int) {
	log.Printf("Wait %v then re-connnect. retry times: %d", c.RetryInterval, times)
	if times > 0 {
		defer func() {
			times -= 1
			c.ConsumerReConnect(times)
		}()
	}

	time.Sleep(c.RetryInterval)
	c.Consumer()
}

// Publish This a method is called to publish message into queue
func (c *Connection) Publish(message string) error {
	conn, err := amqp.Dial(c.Url)
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		c.Queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	// exchange is empty， use the default exchange of rabbitMQ
	// else declare exchange and bind queue on it
	// declare queue and bind it to exchange
	if len(c.Exchange) > 0 {
		err = ch.ExchangeDeclare(
			c.Exchange,
			c.ExchangeType,
			true,
			false,
			false,
			false,
			nil)
		if err != nil {
			return nil
		}

		err = ch.QueueBind(q.Name, q.Name, c.Exchange, true, nil)
		if err != nil {
			return nil
		}
	}

	if c.PublishTimeout == 0 {
		c.PublishTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.PublishTimeout)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		c.Exchange, // exchange
		c.Queue,    // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return err
	}

	return nil
}
