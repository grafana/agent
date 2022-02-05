# Use the loki-build-image as a base as it shares most of the common tooling.
# We just need some additional tooling for building packages.
FROM grafana/loki-build-image:0.18.0

RUN apt-get update && apt-get install -y \
  build-essential \
  rpm \
  ruby \
  ruby-dev \
  rubygems \
  && gem install --no-ri --no-rdoc fpm \
  && rm -rf /var/lib/apt/lists/*

# Install dependencies used for the operator
# Keep versions in sync with cmd/agent-operator/DEVELOPERS.md
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0

# Fix permissions for /go/bin directory.
RUN chmod 0755 /go/bin
# Use /src/agent directory instead of /src/loki.
RUN sed -i -e 's /src/loki /src/agent ' /build.sh
