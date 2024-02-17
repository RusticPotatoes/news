# Dockerfile
FROM golang:1.16-buster AS build
# Enable Go modules
ENV GO111MODULE=on
RUN go install github.com/go-delve/delve/cmd/dlv@v1.7.1
COPY . /src/news
WORKDIR /src/news
RUN CGO_ENABLED=0 go build -o news 

FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates nodejs npm

COPY --from=build /src/news/readability-server /app/readability-server

WORKDIR /app/readability-server

RUN npm install

WORKDIR /app

COPY --from=build /src/news/news /app/
COPY --from=build /src/news/static /app/static
COPY --from=build /src/news/tmpl /app/tmpl
EXPOSE 8080 40000
ENTRYPOINT ["/go/bin/dlv", "exec", "/app/news", "--continue", "--accept-multiclient", "--api-version=2", "--headless", "--listen=:40000"]