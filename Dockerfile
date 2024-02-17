# Dockerfile
FROM golang:1.20-buster AS build
# Enable Go modules
ENV GO111MODULE=on

COPY . /src/news
WORKDIR /src/news
RUN go mod download
RUN go mod vendor
RUN go build -o init-db sql/init-db.go
RUN CGO_ENABLED=0 go build -o news 

FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates nodejs npm
RUN apk add coreutils # Required for sha256sum

COPY --from=build /src/news/readability-server /app/readability-server

WORKDIR /app/readability-server

RUN npm install

WORKDIR /app

COPY --from=build /src/news/news /app/
COPY --from=build /src/news/static /app/static
COPY --from=build /src/news/tmpl /app/tmpl

EXPOSE 8080
CMD ["/app/news"]