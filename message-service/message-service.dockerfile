FROM alpine:latest

RUN mkdir /app

COPY messagingApp /app

CMD [ "/app/messagingApp"]