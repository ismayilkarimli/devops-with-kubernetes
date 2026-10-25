FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum ping.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOARCH=amd64 go build -o server ping.go

FROM scratch

COPY --from=build /app/server /server

ENTRYPOINT [ "/server" ]
