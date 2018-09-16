package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

/*
WatchGameUpdates will watch the Kafka Topic for game start
When the game is started it informs the client
*/
func WatchGameUpdates() {
	// Listen to Kafka Topic

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka",
		"group.id":          os.Getenv("HOSTNAME"),
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{gameUpdatesTopic}, nil)

	log.Printf("Subscribed to Topic " + gameUpdatesTopic)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			var message EventMessage
			json.Unmarshal(msg.Value, &message)
			if sock, ok := allSockets[message.To]; ok {
				sock.WriteJSON(message)
			}

		} else {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
			break
		}
	}

	c.Close()

}

// ConnectProducer connects to the kafka producer
func ConnectProducer() *kafka.Producer {

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka",
	})

	// Delivery report handler for produced messages
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	producer.Flush(15 * 1000)

	if err != nil {
		panic("Cannot connect ")
	}

	return producer
}

var gameUpdatesTopic = "gameUpdates"

// EventMessage defines an event coming through the socket
type EventMessage struct {
	Event   EventName
	Payload interface{}
	To      int
}

// Send sends a message to the Kafka Topic
func (e EventMessage) Send(producer *kafka.Producer) {
	gameUpdateMessage, err := json.Marshal(e)

	if err != nil {
		log.Fatalf("Cannot send message to topic %s", err.Error())
		return
	}

	producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &gameUpdatesTopic,
			Partition: kafka.PartitionAny,
		},
		Value: gameUpdateMessage,
	}, nil)
}
