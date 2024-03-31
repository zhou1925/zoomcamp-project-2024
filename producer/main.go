package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"bufio"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
)

var producer *kafka.Producer
var mutex sync.Mutex

func ReadConfig() kafka.ConfigMap {
	m := make(map[string]kafka.ConfigValue)

	file, err := os.Open("client.properties")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %s", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			kv := strings.Split(line, "=")
			parameter := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[parameter] = value
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s", err)
		os.Exit(1)
	}

	return m
}

type MessageData struct {
	Insights struct {
		TopJobSkills   map[string]int `json:"top_job_skills"`
		TopCompanies   map[string]int `json:"top_companies"`
		TopCities      map[string]int `json:"top_cities"`
		TopStates      map[string]int `json:"top_states"`
	} `json:"insights"`
}

func ProduceMessage(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	var messageData MessageData
	if err := c.BindJSON(&messageData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic := "test_simple"
	valueJSON, err := json.Marshal(messageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          valueJSON,
	}, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}

func main() {
	// Set up Gin router
	router := gin.Default()

	// Initialize Kafka producer
	conf := ReadConfig()
	p, err := kafka.NewProducer(&conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kafka producer: %v", err)
		os.Exit(1)
	}
	producer = p

	// Handle produce message endpoint
	router.POST("/produce", ProduceMessage)

	// Run Gin server
	err = router.Run(":8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v", err)
		os.Exit(1)
	}
}
