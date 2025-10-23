# Go Kafka Producer & Consumer

This repository contains a **Kafka producer and consumer** implemented in **Go** using **Confluent Kafka** and **Schema Registry** with **Protobuf serialization**.

---



## Prerequisites

- Go 1.21+
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` plugin:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```
> **Note:** Make sure the Go binaries path is in your `PATH` so `protoc-gen-go` can be found:

````bash
export PATH=$PATH:$(go env GOPATH)/bin
````

---

## Setup

1. **Clone the repository**

```bash
git clone <repo-url>
cd go_confluent_kafka_app
````

2. **Set environment variables**

Copy the sample `.env` files for both producer and consumer:

```bash
cp producer/.env.sample producer/.env
cp consumer/.env.sample consumer/.env
```

Edit `.env` files and provide your Confluent Cloud credentials, topics, and other configuration:

```dotenv
BOOTSTRAP_SERVERS=pkc-xxxx.us-east-2.aws.confluent.cloud:9092,pkc-xxxx.us-east-2.aws.confluent.cloud:9093
TOPIC_NAMES=test-topic
KAFKA_API_KEY=<your-kafka-api-key>
KAFKA_API_SECRET=<your-kafka-api-secret>
SCHEMA_REGISTRY_URL=https://psrc-xxxx.us-east-2.aws.confluent.cloud
SCHEMA_REGISTRY_API_KEY=<your-sr-key>
SCHEMA_REGISTRY_API_SECRET=<your-sr-secret>
CONSUMER_GROUP_ID=go-kafka-group
CONSUMER_AUTO_OFFSET=earliest
```

> ⚠️ Make sure both producer and consumer `.env` files are updated.

---

## Generate Protobuf Code

For both producer and consumer:

```bash
cd producer
make proto
cd ../consumer
make proto
```

This generates `.pb.go` files from `user.proto`.

---

## Build

Build both producer and consumer:

```bash
# From repo root
make build-producer
make build-consumer
```

Binaries are created in `bin/producer` and `bin/consumer`.

---

## Run

Run producer or consumer:

```bash
# Producer
make run-producer

# Consumer
make run-consumer
```

> The apps will read configuration from `.env`.

---

## Clean

To remove binaries:

```bash
make clean
```

---

## Notes

- **Schema Registry**: Ensure the Protobuf schemas are registered in the schema registry if producing for the first time.
- **Topics**: Make sure your Kafka topics exist and your credentials have read/write access.
- **Environment Variables**: Never commit real credentials to Git. Use `.env.sample` for reference.

---

## Optional Architecture Diagram

```
Producer (Go) ---> Kafka Topic ---> Consumer (Go)
        \                         /
         \-- Schema Registry ----/
```

- Messages are serialized with **Protobuf**
- Schema IDs are embedded in messages for automatic deserialization
