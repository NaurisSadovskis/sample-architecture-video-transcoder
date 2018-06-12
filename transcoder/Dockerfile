FROM ubuntu:18.04

RUN apt-get update && apt-get install \
     ffmpeg libswsca* libavfor* libavre* -y

# This file must be already built.
ADD transcoder-service /transcoder
ENTRYPOINT ["/transcoder"]