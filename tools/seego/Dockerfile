FROM rfratto/seego:latest as build

# Use custom Go version instead of one prepacked in seego
ENV GOLANG_VERSION 1.17.6
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 231654bbf2dab3d86c1619ce799e77b03d96f9b50770297c8f4dff8836fc8ca2
RUN  rm -rf /usr/local/go                                           \
  && curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz             \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz                           \
  && rm golang.tar.gz
RUN apt-get update
RUN apt-get install -y nsis
