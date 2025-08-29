package kafka_listener

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	"web_service/internal/domain"
	"web_service/internal/usecase"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer         *kafka.Consumer
	producer         *kafka.Producer
	saveOrderUseCase *usecase.SaveOrderUseCase
}

func NewConsumer(brokers, groupID string, saveUseCase *usecase.SaveOrderUseCase) (*Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, err
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		cErr := consumer.Close()
		if cErr != nil {
			return nil, cErr
		}
		return nil, err
	}
	return &Consumer{
		consumer:         consumer,
		producer:         producer,
		saveOrderUseCase: saveUseCase,
	}, nil
}

func (c *Consumer) Subscribe(topic string) error {
	return c.consumer.Subscribe(topic, nil)
}

func (c *Consumer) Close() {
	c.producer.Close()
	err := c.consumer.Close()
	if err != nil {
		return
	}
}

func (c *Consumer) commitMessage(msg *kafka.Message) {
	_, err := c.consumer.CommitMessage(msg)
	if err != nil {
		log.Printf("Failed to commit offset: %v\n", err)
	} else {
		log.Printf("Committed offset for topic %s partition %d offset %d\n",
			*msg.TopicPartition.Topic,
			msg.TopicPartition.Partition,
			msg.TopicPartition.Offset)
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {

		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Consumer error: %v (%v)\n", err, msg)
				continue
			}
			var order domain.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to parse message: %v\n", err)
				c.sendToDLQ(msg, err)
				c.commitMessage(msg)
				continue
			}
			if err := c.saveOrderUseCase.Save(&order); err != nil {
				if errors.Is(err, domain.OrderAlreadyExistsError) {
					c.commitMessage(msg)
					continue
				}
				log.Printf("Failed to save order: %v\n", err)
				continue
			}
			c.commitMessage(msg)
			log.Printf("Successfully processed order: %s\n", order.OrderUID)
		}

	}
}

func (c *Consumer) sendToDLQ(msg *kafka.Message, parseErr error) {
	dlqMessage := map[string]interface{}{
		"original_message": string(msg.Value),
		"error":            parseErr.Error(),
		"timestamp":        time.Now().Format(time.RFC3339),
		"topic":            *msg.TopicPartition.Topic,
		"partition":        msg.TopicPartition.Partition,
		"offset":           msg.TopicPartition.Offset,
	}

	dlqData, _ := json.Marshal(dlqMessage)

	dlqTopic := "orders_dlq"
	err := c.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &dlqTopic,
			Partition: kafka.PartitionAny,
		},
		Value: dlqData,
	}, nil)

	if err != nil {
		log.Printf("Failed to send to DLQ: %v\n", err)
	}
}
