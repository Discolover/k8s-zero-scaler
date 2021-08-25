FROM golang:1.16-alpine AS build-stage

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./

RUN go build -o zero-scaler

FROM alpine:latest
EXPOSE 8080
COPY --from=build-stage /app/zero-scaler /zero-scaler

ENTRYPOINT ["/zero-scaler"]
