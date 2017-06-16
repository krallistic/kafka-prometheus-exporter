# kafka-prometheus-consumer-exporter

Simple Prometheus exporter to export Metrics around Offsets & Consumergroup. Currently it exports three types of Metrrics:

 - `kafka_broker_current_offset` Current Offset at the Broker by Topic/Partion
 - `kafka_topic_oldest_offset` Oldest Offset still current in topic
 - `kafka_topic_in_sync_replica` Number of replicas which are inSync with the Topic/Partition leader.
 - `kafka_topic_under_replicated` Bool if all replicas are inSync by topic/partition.

 If you use old consumer (which store their offset in zookeeper), these two Metrics about consumergroups and their offset are also aviable:
 - `kafka_consumergroup_current_offset` Current offset of every consumergroup by Topic/Partition
 - `kafka_consumergroup_lag` Approximate Lag of every consumergroup by Topic/Partition. Please note the Lag is not 100% Correct.

# Build 

1. Ensure Dependecies with `dep ensure`
2. Build go binary `go build -o kafka_exporter main/cmd.go`
 
# Usage

Build the Sourcecode and start the binary with the correct command line options.

### Example Output:
```apple js
# HELP kafka_broker_current_offset Current Offset of a Broker at Topic/Partition
# TYPE kafka_broker_current_offset gauge
kafka_broker_current_offset{cluster="kafkaCluster",partition="0",topic="test"} 721271
kafka_broker_current_offset{cluster="kafkaCluster",partition="0",topic="test-verf"} 704461
# HELP kafka_consumergroup_current_offset Current Offset of a ConsumerGroup at Topic/Partition
# TYPE kafka_consumergroup_current_offset gauge
kafka_consumergroup_current_offset{cluster="kafkaCluster",consumergroup="console-consumer-13363",partition="0",topic="test"} 720453
kafka_consumergroup_current_offset{cluster="kafkaCluster",consumergroup="console-consumer-2265",partition="0",topic="test"} 16
kafka_consumergroup_current_offset{cluster="kafkaCluster",consumergroup="console-consumer-45639",partition="0",topic="test-verf"} 704461
kafka_consumergroup_current_offset{cluster="kafkaCluster",consumergroup="console-consumer-58950",partition="0",topic="test"} 380475
# HELP kafka_consumergroup_lag Current Approximate Lag of a ConsumerGroup at Topic/Partition
# TYPE kafka_consumergroup_lag gauge
kafka_consumergroup_lag{cluster="kafkaCluster",consumergroup="console-consumer-13363",partition="0",topic="test"} 818
kafka_consumergroup_lag{cluster="kafkaCluster",consumergroup="console-consumer-2265",partition="0",topic="test"} 721255
kafka_consumergroup_lag{cluster="kafkaCluster",consumergroup="console-consumer-45639",partition="0",topic="test-verf"} 0
kafka_consumergroup_lag{cluster="kafkaCluster",consumergroup="console-consumer-58950",partition="0",topic="test"} 340796
```

# Command line Options:

| Argument | Description | Default |
| --- | --- | --- |
| `listen-address` | The address on which to expose the web interface and generated Prometheus metrics. | `:8080`
| `telemetry-path` | Path under which to expose metrics. | `/metrics`
| `zookeeper-connect` | Zookeeper connection string | `localhost:2181`
| `cluster-name` | Name of the Kafka cluster used in static label |`kafka-cluster` 
| `refresh-interval` | Seconds to sleep in between refreshes | `15`
