FROM golang:alpine
MAINTAINER  Jakob Karalus <jakob.karalus@gnx.net>

ADD kafka_exporter /bin/usr/sbin/kafka_exporter

CMD ["/bin/usr/sbin/kafka_exporter"]