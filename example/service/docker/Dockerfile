FROM registry.dev.minus5.hr/base_go_freetds:20160712094343.b323ed6

ADD service /bin/

EXPOSE 12345
ENV SERVICE_12345_IGNORE 1

WORKDIR /bin
ENTRYPOINT ["service", "-port", "12345"]


