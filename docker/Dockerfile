FROM golang:latest

RUN mkdir /netcache
ADD . /netcache/
WORKDIR /netcache

RUN ls
RUN export GO111MODULE=on
RUN git clone https://github.com/Nigel2392/netcache.git /netcache/github-repo
RUN cd /netcache/github-repo && go get github.com/Nigel2392/go-datastructures
RUN cd /netcache/github-repo && go build -tags="docker" -o /netcache/netcache_server /netcache/github-repo/netcache

EXPOSE 8080

ENTRYPOINT [ "/netcache/netcache_server" ]
# CMD [ "/netcache/netcache_server" ]

