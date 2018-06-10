package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/av/transcode"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
	"github.com/streadway/amqp"
)

func transcodeVideo(b []byte) {
	url := fmt.Sprintf("%s", b)

	fmt.Printf("Received a request to convert: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	epoch := time.Now().Unix()

	out, err := os.Create(strconv.Itoa(int(epoch)) + "_source.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	fmt.Printf("Downloading video locally: %s\n", url)
	io.Copy(out, resp.Body)

	format.RegisterAll()

	fmt.Printf("Starting conversion for: %s\n", url)
	file, _ := avutil.Open(strconv.Itoa(int(epoch)) + "_source.mp4")
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
				fmt.Printf("Decoding in process...", frame.SampleCount)
			}
		}
	}

	trans := &transcode.Demuxer{
		Options: transcode.Options{},
		Demuxer: file,
	}

	outfile, _ := avutil.Create(strconv.Itoa(int(epoch)) + "_transcode.mp4")
	avutil.CopyFile(outfile, trans)

	outfile.Close()
	file.Close()
	trans.Close()

	fmt.Println("Conversion finished! Waiting for more!")
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
		true,   // auto-ack
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
			transcodeVideo(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	// connect to queue, get the message and transcode!

}
