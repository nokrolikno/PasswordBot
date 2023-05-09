FROM golang:latest

COPY ./ ./
ENV GOPATH=/
RUN go mod init github.com/nokrolikno/PasswordBot
RUN go mod tidy
RUN go build -o main ./cmd/bot/
