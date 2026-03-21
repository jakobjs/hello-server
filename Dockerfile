FROM golang:1.26-trixie AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go vet -v
RUN go test -v
RUN go build -o /app/hello-server

FROM gcr.io/distroless/static-debian13
COPY --from=build /app/hello-server /app/hello-server
ENTRYPOINT ["/app/hello-server"]
