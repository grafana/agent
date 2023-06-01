FROM --platform=linux/amd64 ubuntu:22.04 as builder

RUN apt-get update && apt-get -y install gcc make

ADD src.c lib.c docker.sh ./
RUN bash docker.sh

FROM scratch
COPY --from=builder elf elf.debug elf.stripped elf.debuglink elf.nopie elf.nobuildid libexample.so ./elfs/
COPY --from=builder /usr/lib/debug/ ./usr/lib/debug/
