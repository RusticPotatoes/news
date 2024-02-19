# use cgo buster container
FROM golang:1.20.5-buster AS build
# Enable Go modules
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# copy project and download dependencies
COPY . /src/news
WORKDIR /src/news
RUN go mod download
RUN go mod vendor

# build the project
RUN CGO_ENABLED=1 go build -o news 

# use node buster slim container
FROM node:20-buster-slim AS final

# get package dependencies
RUN apt-get update && apt-get install -y sqlite3 && rm -rf /var/lib/apt/lists/*

# COPY --from=build /src/news/readability-server /app/readability-server

# WORKDIR /app/readability-server

# RUN npm install

WORKDIR /app

COPY --from=build /src/news/news /app/
COPY --from=build /src/news/static /app/static
COPY --from=build /src/news/tmpl /app/tmpl

# Create the /app/data directory
RUN mkdir -p /app/data

# init sql db
COPY ./sql/init.sql /app/sql/

RUN chmod +x /app/news
RUN ldd /app/news || true
RUN sqlite3 /app/data/news.db < /app/sql/init.sql

EXPOSE 8080

# Start processes
CMD ["/usr/bin/supervisord"]