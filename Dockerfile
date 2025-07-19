FROM alpine:latest

WORKDIR /app

COPY server /app

EXPOSE 8080
CMD ["./server"]