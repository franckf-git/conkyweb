FROM golang:alpine AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app
# RUN go mod download # if modules
RUN go build -o main main.go

FROM alpine
COPY --from=builder /app /app
WORKDIR /app
CMD ["./main"]