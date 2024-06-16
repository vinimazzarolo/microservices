package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8000", nil)

	fmt.Println("Starting consumer")

	config := &kafka.ConfigMap{
		"bootstrap.servers": "broker:9092",
		"group.id":          "generic-consumer-group",
		"auto.offset.reset": "earliest",
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	defer consumer.Close()

	err = consumer.Subscribe("generic", nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s", err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true

	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			msg, err := consumer.ReadMessage(-1)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			} else {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}
}
