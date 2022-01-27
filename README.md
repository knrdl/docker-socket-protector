# Traefik Docker Protector
Limit traefik's control over the docker daemon

Traefik has a great [docker integration](https://doc.traefik.io/traefik/providers/docker/)! But exposing the docker socket to traefik equals basically giving traefik **full root access** to the host system. This litte program acts as a filtering proxy so traefik gets readonly access to necessary information from docker. See also https://doc.traefik.io/traefik/providers/docker/#endpoint

```
.---------.                    .----------------.                          .--------.
|         |                    | Traefik Docker |                          | Docker |
| Traefik |<--Docker Network-->| Protector      |<--/var/run/docker.sock-->| Daemon |
'---------'                    '----------------'                          '--------'

```

## Setup

```yaml
version: '3.9'

services:

  traefik:
    image: traefik
    command: "--providers.docker.endpoint=http://traefik-docker-protector:2375"
    ports:
      - "80:80"
    networks:
      - docker_socket_net
  
  traefik-docker-protector:
    image: knrdl/traefik-docker-protector
    hostname: traefik-docker-protector
    read_only: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - docker_socket_net

networks:
  docker_socket_net:
    attachable: false
    internal: true
```

## FAQ

### Why not just mount the docker socket as read only? 
Mounting as `/var/run/docker.sock:/var/run/docker.sock:ro` (**ro** = readonly) just prevents traefik from changing file permissions on the socket file. The socket as pipe object stays writable, so you can still send arbitrary requests to the socket. Nevertheless using **ro** mode for socket mount is not wrong, but won't solve the security problem!
