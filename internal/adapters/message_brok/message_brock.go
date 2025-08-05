package messagebrok

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
	topic    string
}

func NewConsumer(brokers []string, topic string) *Consumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka consumer: %v", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    topic,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, message []byte) error) error {
	partitionConsumer, err := c.consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			if err := handler(ctx, msg.Value); err != nil {
				log.Printf("Ошибка обработки сообщения: %v", err)
			}
		case err := <-partitionConsumer.Errors():
			log.Printf("Ошибка Kafka: %v", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type NoopConsumer struct{}

func NewNoopConsumer() *NoopConsumer {
	return &NoopConsumer{}
}

func (c *NoopConsumer) Consume(ctx context.Context, handler func(ctx context.Context, message []byte) error) error {
	log.Println("NoopConsumer: Kafka is disabled, no messages will be consumed")
	<-ctx.Done()
	return ctx.Err()
}
