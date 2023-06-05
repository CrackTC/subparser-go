CGO_ENABLED=0 go build -o main -v
docker buildx build -t cracktc/subparser-go .
docker push cracktc/subparser-go
