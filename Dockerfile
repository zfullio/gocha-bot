FROM golang:alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o gocha /build/cmd/gocha/

FROM alpine
RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow
WORKDIR /app
COPY --from=builder /build/gocha ./
ENTRYPOINT ["./gocha"]