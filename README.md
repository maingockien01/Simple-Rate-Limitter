# Introduction

Welcome to my project of Rate Limiting!

# Run

## Dependencies
- Docker
- docker-compose
- Golang (if you want to run without Docker, optional)

## Build Docker images

There are 3 images:
- Backend:
```
$ docker pull containous/whoami
```
- Rate limiter & Ruler worker:
```
# on main folder

$ chmod +x ./build-image.sh
$ ./build-image.sh
```

## Run
I use docker compose to set up the system:
```
$ docker-compose -f docker-compose.yml up
```

After all Docker containers are up and running, you can access the network through `127.0.0.1:8081` and `127.0.0.1:8080`.

## Run test
For testing `/` path:
```
curl --parallel --parallel-immediate --parallel-max 6 --config urls.txt -i
```

For testing `/whoami` path:
```
curl --parallel --parallel-immediate --parallel-max 6 --config urls-whoami.txt -i 
```