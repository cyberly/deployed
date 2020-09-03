FROM golang:1.13.10-alpine as builder

RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates \
    && apk add git

WORKDIR $GOPATH/src/github.com/cbyerly/deployed
COPY . .

RUN go get -d ./...
RUN go build -o /deployed main.go

FROM alpine:latest
RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates

ENTRYPOINT [ "/deployed" ]
COPY --from=builder /deployed /