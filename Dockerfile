# Build application
FROM golang:1.22-alpine as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /go/bin/server

# Run application on distroless image
FROM gcr.io/distroless/static-debian12:latest-arm64 as runner
WORKDIR /app
COPY ./config.yml .
COPY ./slocraft-logo-64.png .
COPY --from=build /go/bin/server .
CMD ["/app/server"]