package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// InsightData represents the structure of the insights data received
type InsightData struct {
	Insights struct {
		TopJobSkills map[string]int `json:"top_job_skills"`
		TopCompanies map[string]int `json:"top_companies"`
		TopCities    map[string]int `json:"top_cities"`
		TopStates    map[string]int `json:"top_states"`
	} `json:"insights"`
}

// Function to handle insertion of insights data into BigQuery
type InsightRow struct {
	Category string `bigquery:"Category"`
	Name     string `bigquery:"Name"`
	Count    int    `bigquery:"Count"`
}

// Function to handle insertion of insights data into BigQuery
func InsertInsightsData(ctx context.Context, client *bigquery.Client, insights InsightData) error {
	var rows []*InsightRow

	// Iterate through the insights data and convert it into rows for insertion
	for category, data := range map[string]map[string]int{
		"top_job_skills": insights.Insights.TopJobSkills,
		"top_companies":  insights.Insights.TopCompanies,
		"top_cities":     insights.Insights.TopCities,
		"top_states":     insights.Insights.TopStates,
	} {
		for key, count := range data {
			row := InsightRow{
				Category: category,
				Name:     key,
				Count:    count,
			}
			rows = append(rows, &row)
		}
	}
	// Log the rows before insertion
	log.Printf("Rows to be inserted: %+v", rows)

	// Create a BigQuery table inserter
	inserter := client.Dataset("jobs_dataset").Table("jobs_table").Inserter()

	// Insert rows into BigQuery
	if err := inserter.Put(ctx, rows); err != nil {
		return err
	}
	return nil
}




// MessageData represents the structure of the message consumed from Kafka
type MessageData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ReadConfig() kafka.ConfigMap {
	// reads the client configuration from client.properties
	// and returns it as a key-value map
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

func main() {
	// Initialize Kafka producer
	conf := ReadConfig()
	p, err := kafka.NewProducer(&conf)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer p.Close()

	// Initialize a BigQuery client
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "copper-stacker-411515")
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Kafka topic
	topic := "test_simple"

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(&conf)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe to Kafka topic(s)
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to Kafka topics: %v", err)
	}

	// Infinite loop to consume messages from Kafka and insert into BigQuery
	for {
		ev := consumer.Poll(100) // Adjust the polling interval as needed
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case *kafka.Message:
			// Unmarshal Kafka message into InsightData struct
			var insights InsightData
			err := json.Unmarshal(e.Value, &insights)
			if err != nil {
				log.Printf("Failed to unmarshal Kafka message: %v", err)
				continue
			}
			log.Printf("Unmarshaled insights: %+v", insights)

			// Insert insights data into BigQuery
			if err := InsertInsightsData(ctx, client, insights); err != nil {
				log.Printf("Failed to insert insights data into BigQuery: %v", err)
				continue
			}
			log.Println("Inserted insights data into BigQuery successfully")
		case kafka.Error:
			log.Printf("Kafka error: %v", e)
		}
	}

	// Wait for any outstanding acknowledgments to be delivered to Kafka
	time.Sleep(5 * time.Second)
}
