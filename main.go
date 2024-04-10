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

	if queueName == "" {
		fmt.Println("No queue name provided. Please provide a queue name using the -q or --queue flag.")
		os.Exit(1)
	}

	conn, err := connectToRabbitMQ(user, password, url, port)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := openChannel(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	queues, err := getQueuesFromManagementAPI(user, password, url)
	fmt.Printf("--> Found %d queues\n", len(queues))
	if err != nil {
		log.Fatal(err)
	}

	for _, q := range queues {
		if queueName == q.Name || queueName == "all" {
			if err := deleteQueue(ch, q.Name); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func connectToRabbitMQ(user, password, url, port string) (*amqp.Connection, error) {
	rabbitmqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, url, port)
	fmt.Printf("--> Connecting to RabbitMQ at %s\n", rabbitmqUrl)
	conn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	return conn, nil
}

func openChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}
	return ch, nil
}

func deleteQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete the queue %s: %v", queueName, err)
	}
	log.Printf("Queue %s deleted successfully", queueName)
	return nil
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
