# Dockerfile
FROM golang:1.20-buster AS build
# Enable Go modules
ENV GO111MODULE=on

COPY . /src/news
WORKDIR /src/news
RUN go mod download
RUN go mod vendor
# RUN go get github.com/RusticPotatoes/news/dao  
# RUN go get github.com/RusticPotatoes/news/handler
# RUN go get github.com/RusticPotatoes/news/model
# RUN go get github.com/RusticPotatoes/news/service
# RUN go get github.com/RusticPotatoes/news/util
# RUN go get github.com/RusticPotatoes/news/idgn
# RUN go get github.com/RusticPotatoes/news/pkg/util
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