package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	webhookURL := os.Getenv("WEBHOOK_URL")
	if natsURL == "" || webhookURL == "" {
		log.Fatal("NATS_URL and WEBHOOK_URL are required")
	}

	conn, err := nats.Connect(natsURL, nats.RetryOnFailedConnect(true), nats.MaxReconnects(-1))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, subject := range []string{"todo_save", "todo_update"} {
		_, err := conn.QueueSubscribe(subject, "broadcasters", func(msg *nats.Msg) {
			resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(msg.Data))
			if err != nil {
				log.Printf("forward %s: %v", msg.Subject, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Printf("forward %s: HTTP %d", msg.Subject, resp.StatusCode)
				return
			}
			log.Printf("forwarded %s: HTTP %d %s", msg.Subject, resp.StatusCode, msg.Data)
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Print("subscribed to todo_save and todo_update in queue group broadcasters")
	select {}
}
