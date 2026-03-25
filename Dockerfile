FROM golang:1.26-alpine AS build

WORKDIR /code

COPY . .
RUN go build -o app cmd/cmd.go

FROM alpine:3.23

WORKDIR /app
COPY --from=build /code/app .

ENTRYPOINT ["/app/app"]
