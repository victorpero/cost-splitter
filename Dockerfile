# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/amex-grocery-splitter-web ./cmd/amex-grocery-splitter-web

FROM alpine:3.20

RUN addgroup -S -g 1000 app && adduser -S -D -H -u 1000 -G app app

COPY --from=build /out/amex-grocery-splitter-web /usr/local/bin/amex-grocery-splitter-web

USER 1000:1000
EXPOSE 8080

ENTRYPOINT ["amex-grocery-splitter-web"]
CMD ["-addr", "0.0.0.0:8080"]
