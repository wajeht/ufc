FROM golang:1.27.3-alpine@sha256:dbaa054703263d06c8b49576667418473b6231f8c0fa69dee103207fdb8e3654 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ufc-web ./cmd/web

FROM alpine:latest@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1000 -S ufc && adduser -S ufc -u 1000 -G ufc

WORKDIR /app

COPY --from=build /app/ufc-web ./ufc-web

USER ufc

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/health || exit 1

CMD ["./ufc-web"]
