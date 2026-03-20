FROM golang:1.26-trixie
WORKDIR /app
COPY . .
RUN go build -o /app/hello-server
ENTRYPOINT ["/app/hello-server"]

