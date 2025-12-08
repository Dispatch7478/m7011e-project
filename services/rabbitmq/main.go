package main

import (
	"log"
	"net/http"
)

// Trigger pipe.

func main() {
	// 1. Initialize RabbitMQ
	rmq, err := Connect()
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %v", err)
	}
	defer rmq.Conn.Close()
	defer rmq.Channel.Close()

	// 2. Start the background consumer
	rmq.StartConsumer()

	// 3. Simple Handler to Trigger a Message
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		msg := "Hello from Go!"
		
		err := rmq.Publish(msg)
		if err != nil {
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}
		
		w.Write([]byte("Message sent: " + msg))
	})

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}