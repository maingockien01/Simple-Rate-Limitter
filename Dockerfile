FROM golang:latest

RUN apt-get update
RUN apt-get install vim --yes

WORKDIR /home/proxy

COPY . /home/proxy

RUN chmod +x ./install-proxy.sh
RUN ./install-proxy.sh

RUN chmod +x ./run-proxy.sh

ENTRYPOINT ["/bin/bash", "-c", "/home/proxy/run-proxy.sh"]
EXPOSE 8080