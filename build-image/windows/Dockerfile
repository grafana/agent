FROM grafana/ci-build-windows

# Install go
RUN powershell -Command choco install golang --version 1.19.4 -y

# Install gcc compiler for windows
RUN powershell -Command choco install mingw --version 12.2.0 -y
RUN powershell -Command choco install make --version 4.3 -y

# Install nodejs and yarn for the web ui react app
RUN powershell -Command choco install nodejs.install --version 19.2.0 -y
RUN powershell -Command choco install yarn --version 1.22.19 -y
