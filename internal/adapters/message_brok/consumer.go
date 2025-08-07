package messagebrok

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type ConsumerGroupHandler struct {
	handlerFunc func(ctx context.Context, message []byte) error
}

func (h ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.handlerFunc(session.Context(), msg.Value)
		if err != nil {
			log.Printf("Ошибка обработки сообщения: %v", err)
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
}

func NewConsumer(brokers []string, groupID, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         topic,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, message []byte) error) error {
	h := ConsumerGroupHandler{handlerFunc: handler}

	for {
		err := c.consumerGroup.Consume(ctx, []string{c.topic}, h)
		if err != nil {
			log.Printf("Ошибка в consumer group: %v", err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}
