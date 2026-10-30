FROM golang:1.19.2-alpine3.15 AS stage

RUN apk --no-cache add gcc musl-dev

WORKDIR /go/src/wallet-kms

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -a -installsuffix cgo -o app .

FROM alpine:3.16.2

RUN apk --no-cache add ca-certificates

WORKDIR /home/

COPY --from=stage /go/src/wallet-kms/app ./