/**
 * Copyright 2022 Confluent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Example function-based high-level Apache Kafka consumer
package main

// consumer_example implements a consumer using the non-channel Poll() API
// to retrieve messages and events.

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file (if present)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, continuing with system environment variables...")
	}

	// Read configuration values from environment variables
	bootstrapServers := os.Getenv("BOOTSTRAP_SERVERS")
	topic_names := os.Getenv("TOPIC_NAMES")
	kafkaAPIKey := os.Getenv("KAFKA_API_KEY")
	kafkaAPISecret := os.Getenv("KAFKA_API_SECRET")
	srURL := os.Getenv("SCHEMA_REGISTRY_URL")
	srAPIKey := os.Getenv("SCHEMA_REGISTRY_API_KEY")
	srAPISecret := os.Getenv("SCHEMA_REGISTRY_API_SECRET")
	group := os.Getenv("CONSUMER_GROUP_ID")
	auto_offset := os.Getenv("CONSUMER_AUTO_OFFSET")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Validate required environment variables
	required := map[string]string{
		"BOOTSTRAP_SERVERS":          bootstrapServers,
		"TOPIC_NAME":                 topic_names,
		"KAFKA_API_KEY":              kafkaAPIKey,
		"KAFKA_API_SECRET":           kafkaAPISecret,
		"SCHEMA_REGISTRY_URL":        srURL,
		"SCHEMA_REGISTRY_API_KEY":    srAPIKey,
		"SCHEMA_REGISTRY_API_SECRET": srAPISecret,
		"CONSUMER_GROUP_ID":          group,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("❌ Missing required environment variable: %s", key)
		}
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           group,
		"security.protocol":  "SASL_SSL",
		"sasl.mechanism":     "PLAIN",
		"sasl.username":      kafkaAPIKey,
		"sasl.password":      kafkaAPISecret,
		"session.timeout.ms": 10000,
		"auto.offset.reset":  auto_offset})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	// Schema registry client
	srAuthUserInfo := fmt.Sprintf("%s:%s", srAPIKey, srAPISecret)
	config := schemaregistry.NewConfig(srURL)
	config.BasicAuthCredentialsSource = "USER_INFO"
	config.BasicAuthUserInfo = srAuthUserInfo

	client, err := schemaregistry.NewClient(config)
	if err != nil {
		fmt.Printf("Failed to create schema registry client: %s\n", err)
		os.Exit(1)
	}

	deser, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())

	if err != nil {
		fmt.Printf("Failed to create deserializer: %s\n", err)
		os.Exit(1)
	}

	// Register the Protobuf type so that Deserialize can be called.
	// An alternative is to pass a pointer to an instance of the Protobuf type
	// to the DeserializeInto method.
	deser.ProtoRegistry.RegisterMessage((&User{}).ProtoReflect().Type())

	topics := strings.Split(topic_names, ",")

	err = c.SubscribeTopics(topics, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topics: %s\n", err)
		os.Exit(1)
	}

	run := true

	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				value, err := deser.Deserialize(*e.TopicPartition.Topic, e.Value)
				if err != nil {
					fmt.Printf("Failed to deserialize payload: %s\n", err)
				} else {
					fmt.Printf("%% Message on %s:\n%+v\n", e.TopicPartition, value)
				}
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}
