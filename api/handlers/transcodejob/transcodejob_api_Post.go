// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package transcodejob

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/NaurisSadovskis/sample-architecture-video-transcoder/api/types"
	"github.com/streadway/amqp"
)

// TODO: This can be done on web layer instead.
func checkVideo(u string) error {
	response, err := http.Get(u)
	if err != nil {
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

// Post is the handler for POST /transcode/job
func (api TranscodejobAPI) Post(w http.ResponseWriter, r *http.Request) {

	// TODO START: Move this out
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
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

	// TODO END: Move this out

	var reqBody types.TranscodeJob

	// decode request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(400)
		return
	}

	err = checkVideo(reqBody.VideoURL)
	if err != nil {
		log.Println(err)
		return
	}

	// Decode it into JSON
	jsonPayload, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Publish to Queue
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonPayload,
		})

	log.Printf("Message published to %s!\n", q.Name)

}
