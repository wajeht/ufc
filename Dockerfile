FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ufc-web ./cmd/web

FROM alpine:latest@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1000 -S ufc && adduser -S ufc -u 1000 -G ufc

WORKDIR /app

COPY --from=build /app/ufc-web ./ufc-web

USER ufc

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/health || exit 1

CMD ["./ufc-web"]
