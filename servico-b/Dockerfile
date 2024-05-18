FROM golang:1.22 as build
WORKDIR /app
COPY . .
WORKDIR /app/cmd/microservice
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudrun

FROM scratch
WORKDIR /app
COPY --from=build /app/cmd/microservice/cloudrun .
COPY cmd/microservice/.env .
ENTRYPOINT [ "./cloudrun" ]