FROM ubuntu
COPY ./packages.list /packages.list
ARG  DEBIAN_FRONTEND=noninteractive  
RUN  dpkg --add-architecture i386 
RUN  apt-get update -y && apt-get install -y $(grep -vE "^\s*#" /packages.list  | tr "\n" " ") -y 
RUN  gem install --no-document fpm 
RUN  mkdir -p /usr/local/go/bin 
RUN  wget https://golang.org/dl/go1.16.5.linux-amd64.tar.gz 
RUN  tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz 
RUN  rm go1.16.5.linux-amd64.tar.gz 
ENV  PATH=$PATH:/usr/local/go/bin 
RUN  apt-get clean && rm -rf /var/lib/apt/lists/* 
RUN  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg 
RUN  echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null 
RUN  apt-get update 
RUN  apt-get install -y docker-ce docker-ce-cli containerd.io 
ARG OSXCROSS_SDK_URL
ENV OSXCROSS_PATH=/usr/osxcross \
    OSXCROSS_REV=035cc170338b7b252e3f13b0e3ccbf4411bffc41 \
    SDK_VERSION=11.1
RUN  mkdir -p /tmp/osxcross && cd /tmp/osxcross                                           \
  && curl -sSL "https://codeload.github.com/tpoechtrager/osxcross/tar.gz/${OSXCROSS_REV}" \
      | tar -C /tmp/osxcross --strip=1 -xzf -                                             \
  && curl -sSLo tarballs/MacOSX${SDK_VERSION}.sdk.tar.xz ${OSXCROSS_SDK_URL}              \
  && UNATTENDED=yes ./build.sh                                                            \
  && mv target "${OSXCROSS_PATH}"                                                         \
  && rm -rf /tmp/osxcross "/usr/osxcross/SDK/MacOSX${SDK_VERSION}.sdk/usr/share/man"
RUN  dpkg --add-architecture amd64    \
  && dpkg --add-architecture i386     \
  && dpkg --add-architecture armel    \
  && dpkg --add-architecture armhf    \
  && dpkg --add-architecture arm64    \
  && dpkg --add-architecture mips     \
  && dpkg --add-architecture mipsel   \
  && dpkg --add-architecture powerpc  \
  && dpkg --add-architecture ppc64el  \
  && dpkg --add-architecture s390x    \
  && apt-get update                   \
  && apt-get install -yq              \
        binfmt-support                \
        binutils-multiarch            \
        binutils-multiarch-dev        \
        build-essential               \
        bzr                           \
        clang                         \
        crossbuild-essential-arm64    \
        crossbuild-essential-armel    \
        crossbuild-essential-armhf    \
        crossbuild-essential-mipsel   \
        crossbuild-essential-powerpc  \
        crossbuild-essential-ppc64el  \
        crossbuild-essential-s390x    \
        curl                          \
        git                           \
        git-core                      \
        libssl-dev                    \
        llvm                          \
        mercurial                     \
        mingw-w64                     \
        openssl                       \
        patch                         \
        qemu-user-static              \
        software-properties-common    \
        subversion                    \
        sudo                          \
        wget                          \
        xz-utils                      \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
RUN echo "deb http://deb.debian.org/debian buster-backports main" >> /etc/apt/sources.list
RUN  apt-get update \
  && apt-get install -t buster-backports -qy libsystemd-dev \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

#
# OSX
#

ENV OSXCROSS_PATH=/usr/osxcross
COPY --from=osxcross $OSXCROSS_PATH $OSXCROSS_PATH
ENV PATH $OSXCROSS_PATH/bin:$PATH


#
# FreeBSD
#

COPY assets/freebsd-amd64-11.3.tar.xz /usr/freebsd/freebsd-amd64.tar.xz
COPY assets/freebsd-i386-11.3.tar.xz /usr/freebsd/freebsd-i386.tar.xz

RUN  mkdir /usr/freebsd/x86_64-pc-freebsd11 \
  && cd /usr/freebsd/x86_64-pc-freebsd11 \
  && tar -xf /usr/freebsd/freebsd-amd64.tar.xz ./lib ./usr/lib ./usr/include \
  && rm /usr/freebsd/freebsd-amd64.tar.xz \
  \
  && mkdir /usr/freebsd/i386-pc-freebsd11 \
  && cd /usr/freebsd/i386-pc-freebsd11 \
  && tar -xf /usr/freebsd/freebsd-i386.tar.xz ./lib ./usr/lib ./usr/include \
  && rm /usr/freebsd/freebsd-i386.tar.xz

CMD ["bash"]