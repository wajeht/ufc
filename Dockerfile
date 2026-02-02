FROM golang:1.23.4-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ufc-web ./cmd/web

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1001 -S ufc && adduser -S ufc -u 1001 -G ufc

WORKDIR /app

COPY --from=build /app/ufc-web ./ufc-web

USER ufc

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/health || exit 1

CMD ["./ufc-web"]
