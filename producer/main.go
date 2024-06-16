package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Data struct {
	Message string `json:"message"`
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/produce", handleProduce)
	http.ListenAndServe(":8000", nil)
}

func handleProduce(w http.ResponseWriter, r *http.Request) {
	var data Data

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	err := produce(data)
	if err != nil {
		http.Error(w, "Failed to produce message", http.StatusInternalServerError)
		return
	}
}

func produce(d Data) error {
	broker := "broker:9092"

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})

	if err != nil {
		return err
	}

	defer p.Close()
	defer p.Flush(15 * 1000)

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Successfully produced record to topic %s partition [%d] @ offset %v\n", *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()

	topic := "generic"

	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte("foo"),
		Value:          []byte(d.Message),
	}, nil)

	return nil
}
