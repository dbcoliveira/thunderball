FROM alpine:latest

RUN apk update
COPY ./thunderball_linux thunderball
RUN chmod 777 ./thunderball
EXPOSE 7337
CMD ["./thunderball"]
