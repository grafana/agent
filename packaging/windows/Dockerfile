FROM debian
RUN apt-get update && \
    apt-get install -y nsis
RUN mkdir app
WORKDIR /home
COPY logo.ico logo.ico
COPY install_script.nsis install_script.nsis
COPY agent-windows-amd64.exe agent-windows-amd64.exe
ENTRYPOINT ["makensis", "-V4", "/home/install_script.nsis"]

