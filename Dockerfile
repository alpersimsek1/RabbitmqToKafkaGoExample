FROM golang:alpine
ADD . /go/src/app
WORKDIR /go/src/app
ENV PORT=3001
RUN apk add git && go get github.com/segmentio/kafka-go && go get github.com/streadway/amqp
CMD ["go", "run", "main.go"]