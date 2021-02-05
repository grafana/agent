FROM rfratto/seego:latest as build

# Use custom Go version instead of one prepacked in seego
ENV GOLANG_VERSION 1.15.8
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 d3379c32a90fdf9382166f8f48034c459a8cc433730bc9476d39d9082c94583b
RUN  rm -rf /usr/local/go                                           \
  && curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz             \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz                           \
  && rm golang.tar.gz
