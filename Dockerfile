FROM alpine:3.6
MAINTAINER  Jakob Karalus <jakob.karalus@gnx.net>

ADD kafka_exporter /bin/usr/sbin/kafka_exporter

CMD ["/bin/usr/sbin/kafka_exporter"]