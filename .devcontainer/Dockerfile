FROM mcr.microsoft.com/devcontainers/go:0-1.19-bullseye

RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    postgresql-client \
    #protobuf-compiler \
    goreleaser

ENV PROTOCZIP=protoc-3.15.8-linux-x86_64.zip
RUN mkdir -p /usr/bin/protoc \
    && curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/$PROTOCZIP \
    && unzip -o $PROTOCZIP -d /usr/local bin/protoc \
    && sudo chmod 755 /usr/local/bin/protoc \
    && unzip -o $PROTOCZIP -d /usr/local 'include/*' \
    && sudo chmod 755 -R /usr/local/include/google \
    && rm -f $PROTOCZIP

# [Optional] Uncomment the next lines to use go get to install anything else you need
# USER vscode
# RUN go get -x <your-dependency-or-tool>
# USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

RUN echo 'vscode ALL=(ALL:ALL) NOPASSWD:ALL' >> /etc/sudoers.d/vscode

USER vscode
RUN go mod download || echo "Go mod download failed"
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest 
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install github.com/favadi/protoc-go-inject-tag@latest
RUN go install -v golang.org/x/tools/cmd/godoc@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1
RUN go install github.com/golang/mock/mockgen@v1.6.0
RUN go install -v github.com/cweill/gotests/gotests@v1.6.0