FROM ubuntu:15.10

RUN apt-get install -y -q wget libsystemd-dev pkg-config

# Go
RUN wget -q -O go.tar.gz https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go.tar.gz

RUN mkdir -p /opt/go

ENV GOROOT /usr/local/go
ENV GOPATH /opt/go

ENTRYPOINT ["/usr/local/go/bin/go"]


