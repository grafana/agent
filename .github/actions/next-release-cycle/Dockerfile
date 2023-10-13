FROM golang:1.21-alpine3.18

COPY . /src
WORKDIR /src

RUN go build -o /bin/next-release-cycle .

ENTRYPOINT ["/bin/next-release-cycle"]
