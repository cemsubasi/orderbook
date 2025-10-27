package event

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *KafkaPublisher) Publish(eventType string, payload any) error {
	data := map[string]any{
		"type":    eventType,
		"payload": payload,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(context.Background(), kafka.Message{
		Value: bytes,
	})
	if err != nil {
		log.Printf("kafka publish err: %v", err)
	}
	return err
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
