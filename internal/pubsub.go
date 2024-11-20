package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Transient SimpleQueueType = iota
	Persistent
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	pubErr := ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if pubErr != nil {
		return pubErr
	}
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
	q, qErr := ch.QueueDeclare(queueName, simpleQueueType == Persistent, simpleQueueType == Transient, simpleQueueType == Transient, false, nil)
	if qErr != nil {
		return ch, amqp.Queue{}, qErr
	}

	ch.QueueBind(q.Name, key, exchange, false, nil)

	return ch, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	_, _, declareBindErr := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if declareBindErr != nil {
		return declareBindErr
	}
	return nil
}
