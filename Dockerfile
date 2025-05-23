FROM docker.io/golang:1.24.3-alpine3.20 as builder

WORKDIR /go/src/app

COPY src .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /docker-socket-protector





FROM scratch

EXPOSE 2375/tcp

COPY --from=builder /docker-socket-protector /docker-socket-protector

COPY profiles /profiles

CMD ["/docker-socket-protector"]
