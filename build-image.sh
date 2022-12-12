docker build . -f Dockerfile -t proxy
docker build . -f Dockerfile -t proxy2 
docker build . -f cmd/ruler/Dockerfile -t ruler