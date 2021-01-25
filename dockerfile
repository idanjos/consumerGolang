#  base image for Go
FROM golang:latest

RUN mkdir /app

# Set the Current Working Directory inside the container
WORKDIR /app

ADD . /app
ENV GOPATH=/app
# Build the Go app
RUN go build virhus
EXPOSE 3003
ENTRYPOINT ["go","run","virhus","/app/src/virhus","85.217.171.67"]
