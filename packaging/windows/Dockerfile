FROM debian
RUN apt-get update && \
    apt-get install -y nsis && \
    apt-get install -y wget && \
    apt-get install -y unzip
RUN mkdir app    
WORKDIR /home
RUN wget https://nssm.cc/release/nssm-2.24.zip
RUN unzip nssm-2.24.zip
RUN wget --secure-protocol=auto --https-only https://github.com/prometheus-community/windows_exporter/releases/download/v0.15.0/windows_exporter-0.15.0-amd64.exe
RUN mv ./nssm-2.24/win64/nssm.exe ./nssm.exe
COPY logo.ico logo.ico
COPY install_script.nsis install_script.nsis
COPY agent-windows-amd64.exe agent-windows-amd64.exe
ENTRYPOINT ["makensis", "-V4", "/home/install_script.nsis"]