package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
)

type QueueInfo struct {
	Name string `json:"name"`
}

func init() {
	flag.Usage = func() {
		h := "Delete local rabbitmq queues:\n"
		h += "Usage:\n"
		h += " qclean [OPTIONS] [QUEUE]\n"
		h += "Options:\n"
		h += "    -h, --help\n"
		h += "    -v, --version\n"
		h += "    -q, --queue Queue name\n"
		h += "    -p --port Port number\n"
		h += "    -u --user User name\n"
		h += "    -ps --password Password\n"
		h += "    -url --url URL\n"
		h += "Example:\n"
		h += " qclean -q queue_name\n"
		fmt.Fprintf(os.Stderr, h)
	}
}

func main() {

	var (
		queueName string
		port      string
		user      string
		password  string
		url       string
	)

	flag.StringVar(&queueName, "q", "", "Queue name")
	flag.StringVar(&queueName, "queue", "", "Queue name")
	flag.StringVar(&port, "p", "5672", "Port number")
	flag.StringVar(&port, "port", "5672", "Port number")
	flag.StringVar(&user, "u", "guest", "User name")
	flag.StringVar(&user, "user", "guest", "User name")
	flag.StringVar(&password, "ps", "guest", "Password")
	flag.StringVar(&password, "password", "guest", "Password")
	flag.StringVar(&url, "url", "localhost", "URL")
	flag.Parse()

	fmt.Println("Queue Name:", queueName)
	// Connect to RabbitMQ server
	rabbitmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, url, port)
	conn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Fetch all existing queues using RabbitMQ Management API
	queues, err := getQueuesFromManagementAPI(user, password, url)
	if err != nil {
		log.Fatalf("Failed to fetch queues from RabbitMQ Management API: %v", err)
	}

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	for _, q := range queues {
		if queueName == q.Name {
			if _, err = ch.QueueDelete(q.Name, false, false, false); err != nil {
				log.Fatalf("Failed to delete the queue %s: %v", q.Name, err)
			}
			log.Printf("Queue %s deleted successfully", q.Name)

			return
		}
		if q.Name != queueName {
			if _, err = ch.QueueDelete(q.Name, false, false, false); err != nil {
				log.Fatalf("Failed to delete the queue %s: %v", q.Name, err)
			}
		}
		log.Printf("Queue %s deleted successfully", q.Name)
	}
}

func getQueuesFromManagementAPI(username string, password string, url string) ([]QueueInfo, error) {
	rabbitmqUrl := fmt.Sprintf("http://%s:%s@%s:15672/api/queues", username, password, url)
	resp, err := http.Get(rabbitmqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var queues []QueueInfo
	err = json.NewDecoder(resp.Body).Decode(&queues)
	if err != nil {
		return nil, err
	}

	return queues, nil
}
