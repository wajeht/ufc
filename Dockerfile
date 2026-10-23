FROM golang:1.27.3-alpine@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ufc-web ./cmd/web

FROM alpine:latest@sha256:71e9b7bae160ba167fa17414401b94cbcdaf3c32feeab9298bd794a2a0614c35

RUN apk --no-cache add ca-certificates curl

RUN addgroup -g 1000 -S ufc && adduser -S ufc -u 1000 -G ufc

WORKDIR /app

COPY --from=build /app/ufc-web ./ufc-web

USER ufc

EXPOSE 80

HEALTHCHECK CMD curl -f http://localhost/health || exit 1

CMD ["./ufc-web"]
