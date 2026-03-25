FROM golang:1.26-alpine AS build

WORKDIR /code

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app cmd/cmd.go

FROM scratch

WORKDIR /app
COPY --from=build /code/app .

ENTRYPOINT ["/app/app"]
