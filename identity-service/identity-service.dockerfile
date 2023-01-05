FROM alpine:latest

RUN mkdir /app

COPY identityApp /app

CMD [ "/app/identityApp"]