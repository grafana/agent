FROM rfratto/seego:latest as build

# Use custom Go version instead of one prepacked in seego
ENV GOLANG_VERSION 1.18
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 e85278e98f57cdb150fe8409e6e5df5343ecb13cebf03a5d5ff12bd55a80264f
RUN  rm -rf /usr/local/go                                           \
  && curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz             \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz                           \
  && rm golang.tar.gz
RUN apt-get update
RUN apt-get install -y nsis
RUN echo "deb [trusted=yes] https://repo.iovisor.org/apt/bionic bionic-nightly main" | tee /etc/apt/sources.list.d/iovisor.list && apt-get update -y && apt-get install -qy libbpfcc-dev

