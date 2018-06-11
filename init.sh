#!/bin/bash

ENVIRONMENT_CONFIG=".env"

echo "Starting Minio container..."
docker-compose up -d 
echo "Sleeping for 5 seconds..."
sleep 5 # time to intialise

if [ ! -f $ENVIRONMENT_CONFIG ]; then
    echo "Writing Minio secrets to $ENVIRONMENT_CONFIG file..."
    docker-compose logs minio | grep AccessKey | sed 's/.*: /MINIO_ACCESS_KEY=/' >> $ENVIRONMENT_CONFIG
    docker-compose logs minio | grep SecretKey | sed 's/.*: /MINIO_SECRET_KEY=/' >> $ENVIRONMENT_CONFIG
    echo "Done!"
fi


