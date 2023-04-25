FROM golang:latest

RUN mkdir /netcache
ADD . /netcache/
WORKDIR /netcache

RUN export GO111MODULE=on
RUN git clone https://github.com/Nigel2392/netcache.git /netcache/github-repo
RUN go build -tags="docker" -o /netcache/netcache_server /netcache/github-repo/netcache/netcache

EXPOSE 2392

# ENTRYPOINT [ "/netcache/netcache_server" ]
CMD [ "/netcache/netcache_server" ]

