FROM library/alpine:latest

RUN apk --update add tzdata ca-certificates
RUN cp /usr/share/zoneinfo/Europe/Zagreb /etc/localtime

ADD bin/* /bin/

WORKDIR /bin
ENTRYPOINT ["math"]
