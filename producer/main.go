package main

import (
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/joho/godotenv"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func main() {
	// Load environment variables from .env file (if present)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, continuing with system environment variables...")
	}

	// Read configuration values from environment variables
	bootstrapServers := os.Getenv("BOOTSTRAP_SERVERS")
	topic := os.Getenv("TOPIC_NAME")
	kafkaAPIKey := os.Getenv("KAFKA_API_KEY")
	kafkaAPISecret := os.Getenv("KAFKA_API_SECRET")
	srURL := os.Getenv("SCHEMA_REGISTRY_URL")
	srAPIKey := os.Getenv("SCHEMA_REGISTRY_API_KEY")
	srAPISecret := os.Getenv("SCHEMA_REGISTRY_API_SECRET")

	// Validate required environment variables
	required := map[string]string{
		"BOOTSTRAP_SERVERS":          bootstrapServers,
		"TOPIC_NAME":                 topic,
		"KAFKA_API_KEY":              kafkaAPIKey,
		"KAFKA_API_SECRET":           kafkaAPISecret,
		"SCHEMA_REGISTRY_URL":        srURL,
		"SCHEMA_REGISTRY_API_KEY":    srAPIKey,
		"SCHEMA_REGISTRY_API_SECRET": srAPISecret,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("❌ Missing required environment variable: %s", key)
		}
	}

	// Create Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"security.protocol": "SASL_SSL",
		"sasl.mechanism":    "PLAIN",
		"sasl.username":     kafkaAPIKey,
		"sasl.password":     kafkaAPISecret,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	fmt.Printf("✅ Created Kafka Producer: %v\n", p)

	// Schema registry client
	srAuthUserInfo := fmt.Sprintf("%s:%s", srAPIKey, srAPISecret)

	config := schemaregistry.NewConfig(srURL)
	config.BasicAuthCredentialsSource = "USER_INFO"
	config.BasicAuthUserInfo = srAuthUserInfo

	client, err := schemaregistry.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create schema registry client: %v", err)
	}

	// Protobuf serializer
	ser, err := protobuf.NewSerializer(client, serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		log.Fatalf("Failed to create serializer: %v", err)
	}

	deliveryChan := make(chan kafka.Event)

	// --- Create Preferences message ---
	prefs := &Preferences{
		Language: "en",
		DarkMode: true,
	}

	// Pack into Any
	extra, err := anypb.New(prefs)
	if err != nil {
		log.Fatalf("Packing error: %v", err)
	}

	// --- Create User message ---
	value := &User{
		Name:           "First user",
		FavoriteNumber: 42,
		FavoriteColor:  "blue",
		ExtraData:      extra,
	}

	// Serialize
	payload, err := ser.Serialize(topic, value)
	if err != nil {
		log.Fatalf("Failed to serialize payload: %v", err)
	}

	// Produce message
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
	}, deliveryChan)
	if err != nil {
		log.Fatalf("Produce failed: %v", err)
	}

	// Wait for delivery report
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf("❌ Delivery failed: %v", m.TopicPartition.Error)
	} else {
		log.Printf("✅ Delivered message to topic %s [%d] at offset %v",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)
}
