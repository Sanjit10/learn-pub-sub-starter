package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"log/slog"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	QueueTypeTransient SimpleQueueType = iota
	QueueTypeDurable
)

var QueueTypeName = map[SimpleQueueType]string{
	QueueTypeTransient:	"transient",
	QueueTypeDurable:	"durable",
}

func (qt SimpleQueueType) String() string {
	return QueueTypeName[qt]
}


func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	marshalled_val, err := json.Marshal(val)
	if err != nil {
		fmt.Printf("Error encoding the value : %s\n", err)
		return err
	}
	message := amqp.Publishing{
		ContentType: "application/json",
		Body: marshalled_val,
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, message)
	if err != nil {
		fmt.Printf("Error encoding the value : %s\n", err)
		return err
	}
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, 
) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("couldn't establish channel : %w", err)	
	}

	durable := queueType == QueueTypeDurable

	newQueue, err := channel.QueueDeclare(
		queueName,
		durable,
		!durable,
		!durable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}
	err = channel.QueueBind(newQueue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue: %w", err)
	}
	return channel, newQueue, nil
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T),
) error {
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	channel, queue, err := DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		queueType,
	)
	if err != nil {
		logger.Error("Failed to declare and bind queue", "queue", queueName, "error", err.Error())
		return err
	}
	logger.Info("Queue declared and bound successfully", "queue", queueName)

	delivery_channel, err := channel.Consume(queue.Name, "",false, false , false, false ,nil)
	if err != nil {
		logger.Error("Failed to declare and bind channel", "queue", queueName, "error", err.Error())
		return err
	}
	

}