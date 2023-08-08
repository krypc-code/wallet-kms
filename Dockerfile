FROM golang:1.19.2-alpine3.15 as stage

ADD ./ /go/src/wallet-kms

WORKDIR /go/src/wallet-kms

RUN go mod download

RUN go mod tidy

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o app .




FROM alpine:3.16.2

RUN apk --no-cache add ca-certificates

WORKDIR /home/

COPY --from=stage /go/src/wallet-kms/app ./