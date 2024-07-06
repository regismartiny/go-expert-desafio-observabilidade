FROM golang:1.22-alpine as build
WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
COPY . .
WORKDIR /app/cmd/microservice
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app

FROM scratch
WORKDIR /app
COPY --from=build /app/cmd/microservice/app .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY cmd/microservice/.env .
ENTRYPOINT [ "./app" ]