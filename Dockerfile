FROM golang:latest AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o module20-binary

FROM alpine:latest
#LABEL version="1.0.0"
#LABEL maintainer="Svetlana Kostianova<test@test.ru>"
WORKDIR /app
COPY --from=builder /build/module20-binary .
ENTRYPOINT [ "/app/module20-binary"]