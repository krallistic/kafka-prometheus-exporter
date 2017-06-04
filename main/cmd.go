package main
import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	kazoo "github.com/krallistic/kazoo-go"
	"github.com/Shopify/sarama"

	"strconv"

	"fmt"
)

var (
	addr              = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	zookeeperConnect = flag.String("zookeeperConnect", "localhost:2181",  "zookeeper Connection String")
	clusterName = flag.String("clusterName", "kafkaCluster", "Name for the Kafka Cluster")
	refreshInterval = flag.Int("refreshInterval", 15, "Refresh every X Seconds")
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
		Namespace: "kafka",
		Subsystem: "consumergroup",
		Name: "current_offset",
		Help: "Current Offset of a ConsumerGroup at Topic/Partition",
		ConstLabels: map[string]string{"cluster": *clusterName},
		},
		[]string{"consumergroup", "topic", "partition"},
	)

	brokerGougeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "broker",
		Name: "current_offset",
		Help: "Current Offset of a Broker at Topic/Partition",
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
	prometheus.MustRegister(brokerGougeVec)

}

func updateOffsets() {
	startTime := time.Now()
	fmt.Println("Updating Stats, Time: ", time.Now())
	groups, err := zookeeperClient.Consumergroups()
	if err != nil {
		panic(err)
	}

	for _, group := range groups {
		offsets, _ := group.FetchAllOffsets()
		for topicName, partitions :=  range offsets {
			for partition, offset := range partitions{
				//TODO dont recreate Labels everytime
				consumerGroupLabels := map[string]string{"consumergroup": group.Name, "topic": topicName, "partition": strconv.Itoa(int(partition))}
				consumergroupGougeVec.With(consumerGroupLabels).Set(float64(offset))
				brokerOffset, err := brokerClient.GetOffset(topicName, partition, sarama.OffsetNewest)
				if err != nil {
					//TODO
					fmt.Println(err)
				}
				brokerLabels := map[string]string{"topic": topicName, "partition": strconv.Itoa(int(partition))}
				brokerGougeVec.With(brokerLabels).Set(float64(brokerOffset))


				
			}
		}
	}
	fmt.Println("Done Update: ", time.Until(startTime))


}

func updateBrokerOffsets() {

}

func main() {
	flag.Parse()


	var err error
	zookeeperClient, err = kazoo.NewKazooFromConnectionString(*zookeeperConnect, nil)
	if err != nil {
		panic(err)
	}

	brokers, err := zookeeperClient.BrokerList()
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	brokerClient, err = sarama.NewClient(brokers, config)



	// Periodically record some sample latencies for the three services.
	go func() {
		for {
			updateOffsets()
			time.Sleep(time.Duration(time.Duration(*refreshInterval) * time.Second))
		}
	}()



	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}