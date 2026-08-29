FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum main.go ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o server main.go

FROM scratch

COPY --from=build /app/server /server

ENTRYPOINT [ "/server" ]