package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Shopify/sarama"
	kazoo "github.com/krallistic/kazoo-go"

	"fmt"
	"strconv"
)

var (
	listenAddress    = flag.String("listen-address", ":8080", "The address on which to expose the web interface and generated Prometheus metrics.")
	metricsEndpoint  = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
	zookeeperConnect = flag.String("zookeeper-connect", "localhost:2181", "Zookeeper connection string")
	clusterName      = flag.String("cluster-name", "kafka-cluster", "Name of the Kafka cluster used in static label")

	refreshInterval  = flag.Int("refresh-interval", 15, "Seconds to sleep in between refreshes")
)

var (
	partitionOffsetDesc = prometheus.NewDesc(
		"kafka_prartion_current_offset",
		"Current Offset of a Partition",
		[]string{"topic", "partition"},
		map[string]string{"cluster": *clusterName},
	)

	consumerGroupOffset = prometheus.NewDesc(
		"kafka_consumergroup_current_offset",
		"",
		[]string{"consumergroup", "topic", "partition"},
		map[string]string{"cluster": *clusterName},
	)

	consumergroupGougeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "kafka",
		Subsystem:   "consumergroup",
		Name:        "current_offset",
		Help:        "Current Offset of a ConsumerGroup at Topic/Partition",
		ConstLabels: map[string]string{"cluster": *clusterName},
	},
		[]string{"consumergroup", "topic", "partition"},
	)
	consumergroupLagGougeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "kafka",
		Subsystem:   "consumergroup",
		Name:        "lag",
		Help:        "Current Approximate Lag of a ConsumerGroup at Topic/Partition",
		ConstLabels: map[string]string{"cluster": *clusterName},
	},
		[]string{"consumergroup", "topic", "partition"},
	)

	topicCurrentOffsetGougeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "kafka",
		Subsystem:   "topic",
		Name:        "current_offset",
		Help:        "Current Offset of a Broker at Topic/Partition",
		ConstLabels: map[string]string{"cluster": *clusterName},
	},
		[]string{"topic", "partition"},
	)

	topicOldestOffsetGougeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "kafka",
		Subsystem:   "topic",
		Name:        "oldest_offset",
		Help:        "Oldest Offset of a Broker at Topic/Partition",
		ConstLabels: map[string]string{"cluster": *clusterName},
	},
		[]string{"topic", "partition"},
	)

	inSyncReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "kafka",
		Subsystem:   "topic",
		Name:        "in_sync_replica",
		Help:        "InSync teplicas for a topic.",
		ConstLabels: map[string]string{"cluster": *clusterName},
	},
		[]string{"topic", "partition"},
	)
)
var zookeeperClient *kazoo.Kazoo
var brokerClient sarama.Client

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(consumergroupGougeVec)
	prometheus.MustRegister(topicCurrentOffsetGougeVec)
	prometheus.MustRegister(topicOldestOffsetGougeVec)

	prometheus.MustRegister(consumergroupLagGougeVec)
}

func updateOffsets() {
	startTime := time.Now()
	fmt.Println("Updating offset stats, Time: ", time.Now())
	groups, err := zookeeperClient.Consumergroups()
	if err != nil {
		fmt.Println("Error reading consumergroup offsets: ", err)
		initClients()
		return
	}

	for _, group := range groups {
		offsets, _ := group.FetchAllOffsets()
		for topicName, partitions := range offsets {
			for partition, offset := range partitions {
				//TODO dont recreate Labels everytime
				consumerGroupLabels := map[string]string{"consumergroup": group.Name, "topic": topicName, "partition": strconv.Itoa(int(partition))}
				consumergroupGougeVec.With(consumerGroupLabels).Set(float64(offset))
				currentOffset, err := brokerClient.GetOffset(topicName, partition, sarama.OffsetNewest)
				oldestOffset, err2 := brokerClient.GetOffset(topicName, partition, sarama.OffsetOldest)
				if err != nil ||  err2 != nil {
					fmt.Println("Error reading offsets from broker for topic, partition: ", topicName, partition, err)
					initClients()
					return
				}
				brokerLabels := map[string]string{"topic": topicName, "partition": strconv.Itoa(int(partition))}
				topicCurrentOffsetGougeVec.With(brokerLabels).Set(float64(currentOffset))
				topicOldestOffsetGougeVec.With(brokerLabels).Set(float64(oldestOffset))

				consumerGroupLag := currentOffset - offset
				consumergroupLagGougeVec.With(consumerGroupLabels).Set(float64(consumerGroupLag))
			}
		}
	}
	fmt.Println("Done updating offset stats in: ", time.Since(startTime))
}

func updateTopics() {
	startTime := time.Now()
	fmt.Println("Updating  topics stats, Time: ", time.Now())
	brokerClient.Topics()
	fmt.Println("Done updating topics stats in: ", time.Since(startTime))

}

func initClients() {

	fmt.Println("Init zookeeper client with connection string: ", *zookeeperConnect)
	var err error
	zookeeperClient, err = kazoo.NewKazooFromConnectionString(*zookeeperConnect, nil)
	if err != nil {
		fmt.Println("Error Init zookeeper client with connection string:", *zookeeperConnect)
		panic(err)
	}

	brokers, err := zookeeperClient.BrokerList()
	if err != nil {
		fmt.Println("Error reading brokers from zk")
		panic(err)
	}

	fmt.Println("Init Kafka Client with Brokers:", brokers)
	config := sarama.NewConfig()
	brokerClient, err = sarama.NewClient(brokers, config)

	if err != nil {
		fmt.Println("Error Init Kafka Client")
		panic(err)
	}
	fmt.Println("Done Init Clients")
}

func main() {
	flag.Parse()

	//Init Clients
	initClients()

	// Periodically record stats from Kafka
	go func() {
		for {
			updateOffsets()
			time.Sleep(time.Duration(time.Duration(*refreshInterval) * time.Second))
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle(*metricsEndpoint, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
