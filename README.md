# Sample architecture: Video transcoder service

This is a sample architecture consisting for implementing a sample event-driven video transcoding system architecture. System is self-contained and implemented in Docker with no dependencies (apart from Docker) required on the host. 

Both, `api` and `transcoder` are written in Go and use:
* RabbitMQ as a massage broker
* Minio as a storage layer

## Flow

* User submits a request to API for this video to be converted
* API creates a JSON object an passes it as a message to `transcode-requests` RabbitMQ queue
* Transcoder service subscribes to `transcode-requests` queue
* Transcoder service creates a new bucket for the user (todo)
* Trancsoder service transcodes and uploads the final video to the user's minio bucket

## Usage

```
$ curl -s -w  "%{http_code}\n"  -d  '{"videoURL":"https://www.sample-videos.com/video/mp4/720/big_buck_bunny_720p_2mb.mp4", "userID":"amanda5"}' -H "Content-Type: application/json" -X POST http://localhost:5000/transcode/job
```

TODO:
- ~~Vendor Go dependencies~~
- Pass RabbitMQ object to routes to avoid re-initialisation
- Parameterise configuration
- ~~Build API with OpenAPI/RAML~~
- Containerise applications
- Add build steps
- Add structured logging
- Add persistance layer
- Auto credential issuing for Minio
- Create a web layer
