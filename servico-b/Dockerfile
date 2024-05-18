FROM golang:1.22 as build
WORKDIR /app
COPY . .
WORKDIR /app/cmd/microservice
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

FROM scratch
WORKDIR /app
COPY --from=build /app/cmd/microservice/app .
COPY cmd/microservice/.env .
ENTRYPOINT [ "./app" ]