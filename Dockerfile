FROM alpine:latest

WORKDIR /app

COPY main .

EXPOSE 3000

CMD ["./main"]
