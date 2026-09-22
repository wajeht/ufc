FROM golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ufc-web ./cmd/web

FROM alpine:latest@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1000 -S ufc && adduser -S ufc -u 1000 -G ufc

WORKDIR /app

COPY --from=build /app/ufc-web ./ufc-web

USER ufc

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/health || exit 1

CMD ["./ufc-web"]
