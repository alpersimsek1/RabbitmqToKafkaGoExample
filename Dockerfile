FROM golang:alpine
ADD . /go/src/app
WORKDIR /go/src/app

# ADJUST BELOW VARIABLES IN ORDER TO RUN PROCESS
ARG RABBITMQ_URL_ARG=serverUrl
ARG RABBITMQ_PORT_ARG=serverPort

ENV RABBITMQ_URL=$RABBITMQ_URL_ARG
ENV RABBITMQ_PORT=$RABBITMQ_PORT_ARG

ENV PORT=3001
RUN apk add git && go get github.com/segmentio/kafka-go && go get github.com/streadway/amqp
CMD ["go", "run", "main.go"]

