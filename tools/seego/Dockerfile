FROM --platform=$BUILDPLATFORM rfratto/seego:latest as build
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Use custom Go version instead of one prepacked in seego
ENV GOLANG_VERSION 1.15.7
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 0d142143794721bb63ce6c8a6180c4062bcf8ef4715e7d6d6609f3a8282629b3
RUN  rm -rf /usr/local/go                                           \
  && curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz             \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz                           \
  && rm golang.tar.gz
