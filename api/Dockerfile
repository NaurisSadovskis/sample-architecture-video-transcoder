FROM golang:1.10.3-stretch as builder

ADD ./ /go/src/api
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
        cd /go/src/api && \
        dep ensure -v && \
        go build -o transcoder-api *.go  

FROM alpine:3.7  
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/api/transcoder-api /transcoder-api
ENTRYPOINT ["/transcoder-api"]