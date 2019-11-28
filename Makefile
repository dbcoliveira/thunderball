clean:
	rm thunderball
build:
	go mod download
	go build -o thunderball .

docker:
	env GOOS=linux GOARCH=amd64 go build -o thunderball_linux .
	docker build . -t thunderball
	rm thunderball_linux
