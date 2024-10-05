package internal

import (
	"fmt"
	"gateway_service/config"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConsumerEvent struct {
	Key     string
	Message string
	err     error
}

type kafkaHandler struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
}

var kafkahandler *kafkaHandler = &kafkaHandler{
	producer: nil,
	consumer: nil,
}

var cfg *config.Config = config.LoadConfig()

func InitKafka(consumer string) error {
	producerCOnfig := &kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"acks":              "all",
		"retries":           5,
	}

	consumerconfig := &kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"group.id":          consumer,
		"auto.offset.reset": "earliest",
	}

	kproducer, err := kafka.NewProducer(producerCOnfig)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	kconsumer, err := kafka.NewConsumer(consumerconfig)
	if err != nil {
		log.Fatal(err)
		return err
	}

	kafkahandler.consumer = kconsumer

	kafkahandler.producer = kproducer

	return nil
}

func PublishMessage(topic string, key, value []byte) error {
	kDChan := make(chan kafka.Event, 1)

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
	}

	err := kafkahandler.producer.Produce(message, kDChan)
	if err != nil {
		log.Fatal(err)
		return err
	}

	event := <-kDChan

	msg := event.(*kafka.Message)
	if msg.TopicPartition.Error != nil {
		log.Printf("Delivery failed: %v\n", msg.TopicPartition.Error)
		return msg.TopicPartition.Error
	} else {
		fmt.Printf("Message delivered to topic %s [%d] at offset %v\n",
			*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
	}

	close(kDChan)

	return nil
}

func ConsumerTopic(topic string) (messages chan KafkaConsumerEvent) {
	log.Println("Consumer topic was created")
	messages = make(chan KafkaConsumerEvent)
	err := kafkahandler.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			events := kafkahandler.consumer.Poll(100)
			switch e := events.(type) {
			case *kafka.Message:
				fmt.Printf("Message received: %s\n", string(e.Value))
				messages <- KafkaConsumerEvent{Key: string(e.Key), Message: string(e.Value), err: nil}
			case kafka.Error:
				fmt.Printf("Error occurred: %v\n", e)
				messages <- KafkaConsumerEvent{Message: "", err: e}
			}
		}
	}()

	return messages
}
