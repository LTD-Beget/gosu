FROM golang:1.5

RUN mkdir -p /go/src/github.com/LTD-Beget \
	&& git clone https://github.com/LTD-Beget/libcontainer.git /go/src/github.com/LTD-Beget/libcontainer \
	&& cd /go/src/github.com/LTD-Beget/libcontainer \
	&& git checkout --quiet 933ecaadda42dafd3e160258c7c4ac5ddbd9ec07

ENV GOPATH $GOPATH:/go/src/github.com/LTD-Beget/libcontainer/vendor

# disable CGO for ALL THE THINGS (to help ensure no libc)
ENV CGO_ENABLED 0

COPY *.go /go/src/github.com/LTD-Beget/gosu/
WORKDIR /go/src/github.com/LTD-Beget/gosu

# fix uid due to seccomp
RUN groupmod -g 1000 www-data && usermod -u 1000 www-data

# gosu-$(dpkg --print-architecture)
RUN GOARCH=amd64 go build -v -ldflags -d -o /go/bin/gosu-amd64 \
	&& /go/bin/gosu-amd64 www-data id \
	&& /go/bin/gosu-amd64 www-data ls -l /proc/self/fd
