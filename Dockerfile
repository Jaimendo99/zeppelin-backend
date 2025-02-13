FROM alpine:latest

WORKDIR /app

COPY main .

RUN chmod +x main

EXPOSE 3000

CMD ["./main"]
