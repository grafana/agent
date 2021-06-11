FROM debian
RUN apt-get update && \
    apt-get install -y nsis
RUN mkdir app
WORKDIR /home
COPY logo.ico logo.ico
COPY install_script.nsis install_script.nsis
COPY grafana-agent-windows-amd64.exe grafana-agent-windows-amd64.exe
COPY LICENSE LICENSE
ENTRYPOINT makensis -V4 -DVERSION=$VERSION /home/install_script.nsis
