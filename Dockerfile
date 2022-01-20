FROM golang:alpine as builder

WORKDIR /go/src/app

COPY src .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /traefik-docker-protector





FROM scratch

EXPOSE 2375/tcp

COPY --from=builder /traefik-docker-protector /traefik-docker-protector

CMD ["/traefik-docker-protector"]
