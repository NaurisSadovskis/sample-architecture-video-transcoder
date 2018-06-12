
CONTAINER_NAME = "transcoder-build"
CONTAINER_IMG = "golang:1.10.3-stretch"
CONTAINER_ID = $(shell docker inspect --format="{{.Id}}" $(CONTAINER_NAME))

build-docker:
	@echo "Building Dockerfiles..."
	docker-compose build

transcoder-build-start:
ifeq ($(CONTAINER_ID),)
	@echo "Transcoder container not found... Starting!"
	docker run --name $(CONTAINER_NAME) --rm -d -v `pwd`/transcoder:/go/src/transcoder $(CONTAINER_IMG) /bin/bash -c "tail -f /dev/null"
	@echo "Build container started!"
endif

transcoder-binary: transcoder-build-start
	docker exec -it $(CONTAINER_ID) /bin/bash -c "apt-get update && apt-get install \
     ffmpeg libswsca* libavfor* libavre* -y && \
     curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
     cd /go/src/transcoder && dep ensure -v && go build -o transcoder-service *.go"
	 @echo "Build finished!"

start: build-docker
	docker-compose up

stop:
	docker-compose down 
	docker stop $(CONTAINER_ID)