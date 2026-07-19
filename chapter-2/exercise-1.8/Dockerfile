FROM golang:1.25-alpine AS build

WORKDIR /app
COPY main.go .

RUN CGO_ENABLED=0 go build -o server main.go

FROM scratch

COPY --from=build /app/server /server

ENTRYPOINT [ "/server" ]
