# Makefile
setup:
	go install github.com/go-delve/delve/cmd/dlv@v1.20.0
	go mod download
	go mod vendor
	go mod tidy

build:
	docker build --progress=plain -t news-app .

run:
	docker run -p 8080:8080 news-app

buildrun: build
	docker run -p 8080:8080 news-app

up:
	docker-compose up

buildup: build
	docker-compose up --build