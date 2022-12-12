FROM golang:latest

RUN apt-get update
RUN apt-get install vim --yes

WORKDIR /home/proxy

COPY . /home/proxy

RUN chmod +x ./install-ruler.sh
RUN ./install-ruler.sh

RUN chmod +x ./run-ruler.sh

ENTRYPOINT ["/bin/bash", "-c", "/home/proxy/run-ruler.sh"]