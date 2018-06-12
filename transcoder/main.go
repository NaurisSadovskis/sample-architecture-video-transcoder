package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/av/transcode"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
	"github.com/streadway/amqp"
)

type EncodingRequest struct {
	User  string
	Video string
}

func uploadToMinio(user, fname string) error {
	// parameterise ones below
	endpoint := "minio:9000"
	accessKeyID := "S326T87GSXL9K0Y6T6M2"
	secretAccessKey := "bB4Qy2NoLAUAxVug/6pZxM/xsVSlrFnXZcHFLxPC"

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, false)
	if err != nil {
		return err
	}

	err = minioClient.MakeBucket(user, "eu-west-1")
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists("eu-west-1")
		if err == nil && exists {
			log.Printf("We already own %s\n", "eu-west-1")
		} else {
			return err
		}
	}

	// Upload the zip file
	objectName := fname
	filePath := fname
	contentType := "video/mp4"

	// Upload the zip file with FPutObject
	n, err := minioClient.FPutObject(user, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, n)

	os.Remove(fname)

	return nil

}

func transcodeVideo(b []byte) error {
	var m EncodingRequest
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	url := m.Video
	user := strings.ToLower(m.User)

	fmt.Printf("Received a request to convert: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	epoch := time.Now().Unix()

	sourceFn := user + "_" + strconv.Itoa(int(epoch)) + "_source.mp4"
	targetFn := user + "_" + strconv.Itoa(int(epoch)) + "_transcode.mp4"

	out, err := os.Create(sourceFn)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("Downloading video locally: %s\n", url)
	io.Copy(out, resp.Body)

	format.RegisterAll()

	fmt.Printf("Starting conversion for: %s\n", url)
	file, _ := avutil.Open(sourceFn)
	streams, _ := file.Streams()
	var dec *ffmpeg.AudioDecoder

	for _, stream := range streams {
		if stream.Type() == av.AAC {
			dec, _ = ffmpeg.NewAudioDecoder(stream.(av.AudioCodecData))
		}
	}

	for i := 0; i < 10; i++ {
		pkt, _ := file.ReadPacket()
		if streams[pkt.Idx].Type() == av.AAC {
			ok, frame, _ := dec.Decode(pkt.Data)
			if ok {
				fmt.Println("Decoding in process...", frame.SampleCount)
			}
		}
	}

	trans := &transcode.Demuxer{
		Options: transcode.Options{},
		Demuxer: file,
	}

	outfile, _ := avutil.Create(targetFn)
	avutil.CopyFile(outfile, trans)

	outfile.Close()
	file.Close()
	trans.Close()

	os.Remove(sourceFn)

	err = uploadToMinio(user, targetFn)
	if err != nil {
		return err
	}

	return nil

}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Failed to open a conncetion")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel")
	}
	defer ch.Close()

	// Note that we declare the queue here, as well.
	// Because we might start the consumer before the publisher, we want to make sure the queue exists before we try to consume messages from it.
	q, err := ch.QueueDeclare(
		"transcode-requests", // name
		false,                // durable
		false,                // delete when usused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue")
	}

	//  Since it will push us messages asynchronously, we will read the messages from a channel (returned by amqp::Consume) in a goroutine.
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatal("Failed to consume")
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			err := transcodeVideo(d.Body)
			if err != nil {
				d.Ack(false)
			}
			d.Ack(true)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	// connect to queue, get the message and transcode!

}
