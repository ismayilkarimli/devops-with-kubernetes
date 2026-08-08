FROM golang:1.25 AS build

WORKDIR /app

COPY main.go .

RUN CGO_ENABLED=0 go build -o writer main.go

FROM scratch

COPY --from=build /app/writer /writer

ENTRYPOINT [ "/writer" ]