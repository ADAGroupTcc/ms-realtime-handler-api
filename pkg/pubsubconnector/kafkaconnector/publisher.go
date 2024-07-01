package kafkaconnector

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	produceMessage = "produce_message"
	success        = "success"
	failure        = "error"
)

type producerImp struct {
	producer   *kafka.Producer
	logger     *logger.Logger
	instrument interfaces.Instrument
}

func NewKafkaProducer(broker string, log *logger.Logger, instrument interfaces.Instrument) (*producerImp, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})

	if err != nil {
		return nil, err
	}

	producerImpl := producerImp{
		producer:   producer,
		logger:     log,
		instrument: instrument,
	}

	go producerImpl.kafkaEventsReport()

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

	start := time.Now()
	err = p.producer.Produce(kafkaEvent, nil)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("kafka_producer: failed to produce message to producer's queue", err)
		p.sendKafkaMetrics(ctx, produceMessage, failure, config.topic, 0, duration)
		return nil
	}

	p.sendKafkaMetrics(ctx, produceMessage, success, config.topic, 0, duration)

	return err
}

func (p *producerImp) kafkaEventsReport() {
	for event := range p.producer.Events() {
		message := event.(*kafka.Message)
		if message.TopicPartition.Error != nil {
			p.instrument.NoticeError(context.Background(), message.TopicPartition.Error)
			p.logger.Error("kafka_producer: failed to deliver message", message.TopicPartition.Error)
		}
	}
}

func (p *producerImp) sendKafkaMetrics(ctx context.Context, action string, status string, topic string, retryTimes int, duration time.Duration) {
	labels := map[string]interface{}{
		"action":      action,
		"status":      status,
		"topic":       topic,
		"retry_times": retryTimes,
	}

	histogram := p.instrument.StartInt64Histogram(ctx, "chatpicpay_kafka_histogram", "Custom Kafka Producer actions metrics")
	histogram.Add(ctx, int64(duration.Seconds()*1000), labels)
}
