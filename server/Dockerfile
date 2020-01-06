# docker build --rm -t drone/drone .

# FROM drone/ca-certs
FROM r.crun.top/library/golang:1.8
EXPOSE 8000 9000

ENV DATABASE_DRIVER=sqlite3
ENV DATABASE_CONFIG=/var/lib/drone/drone.sqlite
ENV GODEBUG=netdns=go
ENV XDG_CACHE_HOME /var/lib/drone

ADD dist/drone-server /bin/
ADD views /views

CMD ["/bin/drone-server"]
