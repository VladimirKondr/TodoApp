FROM golang:tip-alpine3.23 AS builder

COPY . /go/TodoApp

WORKDIR /go/TodoApp

RUN go mod download 

RUN go build -o /todoapp main.go

FROM alpine

COPY --from=builder /todoapp /todoapp

EXPOSE 8081

ENTRYPOINT ["/todoapp"]

