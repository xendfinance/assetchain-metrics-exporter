build-linux:
	GOARCH=amd64 GOOS=linux go build -o main main.go

build-mac:
	go build -o main main.go  