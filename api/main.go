package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

// size validation can be moved to web tier, instead of API.
func checkVideo(u string) error {
	response, err := http.Get(u)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if response.Header.Get("Content-type") != "video/mp4" {
		return errors.New("This is not a video")
	}

	fileSizeLimit := 10000000 // 10MB

	if response.ContentLength > int64(fileSizeLimit) {
		return errors.New("File is too large. Limit 10MB")
	}

	return nil
}

func main() {
	videoURL := "https://www.sample-videos.com/video/mp4/720/big_buck_bunny_720p_5mb.mp4"

	err := checkVideo(videoURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sending file to RabbitMQ...")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ")
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"transcode-requests",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("Failed to declare a queue")
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(videoURL),
		})

	fmt.Println("Message to RabbitMQ sent!")

}
