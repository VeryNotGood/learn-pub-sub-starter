package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Transient SimpleQueueType = iota
	Persistent
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	fmt.Println("Marshalling the message...")
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	fmt.Println("Successfully marshalled...")
	fmt.Printf("Publishing the message: %v", val)
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	pubErr := ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if pubErr != nil {
		return pubErr
	}
	fmt.Printf("Successfully published...message: %v\nExchange: %v\nKey: %v\n", msg, exchange, key)
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, chErr := conn.Channel()
	if chErr != nil {
		return nil, amqp.Queue{}, chErr
	}

	dlxMap := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	q, qErr := ch.QueueDeclare(queueName, simpleQueueType == Persistent, simpleQueueType == Transient, false, false, dlxMap)
	if qErr != nil {
		return nil, amqp.Queue{}, qErr
	}

	bindErr := ch.QueueBind(q.Name, key, exchange, false, nil)
	if bindErr != nil {
		return nil, amqp.Queue{}, bindErr
	}

	return ch, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	exchangeType,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	ch, _, declareBindErr := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if declareBindErr != nil {
		return declareBindErr
	}
	fmt.Println("Done...")

	deliveryCh, deliveryErr := ch.Consume(queueName, "", false, false, false, false, nil)
	if deliveryErr != nil {
		return deliveryErr
	}

	go func() {
		for delivery := range deliveryCh {
			go func(delivery amqp.Delivery) {
				var message T
				umErr := json.Unmarshal(delivery.Body, &message)
				if umErr != nil {
					fmt.Println(umErr)
					delivery.Nack(false, false)
					return
				}
				ackType := handler(message)

				switch ackType {
				case Ack:
					delivery.Ack(false)
				case NackRequeue:
					delivery.Nack(false, true)
				case NackDiscard:
					delivery.Nack(false, false)
				}
			}(delivery)
		}
	}()

	return nil
}
