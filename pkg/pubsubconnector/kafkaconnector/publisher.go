package kafkaconnector

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type producerImp struct {
	producer *kafka.Producer
}

func NewKafkaProducer(broker string) (*producerImp, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})

	if err != nil {
		return nil, err
	}

	producerImpl := producerImp{
		producer: producer,
	}

	return &producerImpl, nil
}

type publisherConfig struct {
	topic   string
	key     string
	headers map[string]string
}

func getConfigs(configMap *map[string]interface{}) (publisherConfig, error) {
	config := publisherConfig{}
	if configMap == nil {
		return config, errors.New("kafka config must not be null")
	}

	topic, ok := (*configMap)["topic"].(string)
	if !ok {
		return config, errors.New("kafka config.topic must be a string")
	}

	if topic == "" {
		return config, errors.New("kafka config.topic must not be empty")
	}

	config.topic = topic

	partitionKey, ok := (*configMap)["key"].(string)
	if ok {
		config.key = partitionKey
	}

	headers, ok := (*configMap)["headers"].(map[string]string)
	if ok {
		config.headers = headers
	}

	return config, nil
}

func (p *producerImp) Publish(ctx context.Context, message interface{}, configMap *map[string]interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	config, err := getConfigs(configMap)
	if err != nil {
		return err
	}

	kafkaEvent := &kafka.Message{
		Value:          payload,
		Key:            []byte(config.key),
		TopicPartition: kafka.TopicPartition{Topic: &config.topic, Partition: kafka.PartitionAny},
	}

	for key, value := range config.headers {
		kafkaEvent.Headers = append(kafkaEvent.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	err = p.producer.Produce(kafkaEvent, nil)

	if err != nil {
		return nil
	}

	return err
}
