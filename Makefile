# Makefile
build:
	docker build -t news-app .

run: build
	docker run -p 8080:8080 -p 40000:40000 news-app

build-db:
	docker build -t news-db ./sql

run-db: build-db
	docker run -p 5432:5432 news-db

up: build build-db
	docker-compose up