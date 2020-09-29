FROM alpine:3.8

RUN apk add --no-cache \
 bash \
 ruby \
 ruby-dev \
 ruby-etc \
 gcc \
 go \
 git \
 libc-dev \
 libffi-dev \
 make \
 rpm \
 tar \
 && gem install --no-ri --no-rdoc fpm

COPY package.sh /
ENTRYPOINT ["/package.sh"]
