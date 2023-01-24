FROM alpine:latest

RUN mkdir /app

COPY friendApp /app

CMD [ "/app/friendApp"]