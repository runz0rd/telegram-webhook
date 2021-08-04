FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o /app cmd/main.go

FROM scratch
COPY --from=builder /app telegram-webhook
ENTRYPOINT ["/telegram-webhook"]