# Base stage, used to define the base image reference only once and to contain common tools used in other stages.
FROM docker.io/golang:1.24.2-bookworm@sha256:00eccd446e023d3cd9566c25a6e6a02b90db3e1e0bbe26a48fc29cd96e800901 AS base

# Fix: https://github.com/hadolint/hadolint/wiki/DL4006
# Fix: https://github.com/koalaman/shellcheck/wiki/SC3014
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

WORKDIR /opt/app

# Non root user details.
ARG USER_ID=1000
ARG USER_NAME=dev
ARG GROUP_ID=${USER_ID}
ARG GROUP_NAME=${USER_NAME}

RUN \
  # Install tools.
  apt-get update && \
  # TODO: lock_versions to ensure deterministic behaviour.
  apt-get install -y \
  # Install unzip, needed for the protoc protobuf compiler instalation.
  unzip \
  # Install and clang-format, used to format protobuf files.
  # clang-format would ideally be installed from LLVM (https://github.com/llvm/llvm-project/releases/) but that's a ~900MB download and a ~5GB untar so instead it is installed from apt-get.
  clang-format && \
  # Clean up apt update and install unused artifacts.
  apt-get clean && \
  apt-get autoremove && \
  rm -rf /var/lib/apt/lists/* && \
  # Install the protoc protobuf compiler.
  mkdir /tmp/protoc && \
  curl https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip -L -o /tmp/protoc/protoc.zip && \
  unzip /tmp/protoc/protoc.zip -d /tmp/protoc/ && \
  mv /tmp/protoc/bin/protoc /usr/bin/protoc && \
  rm -R /tmp/protoc/ && \
  # Install the Golang protobuf compiler plugin.
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 && \
  # Install the golangci-lint linter.
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.2 && \
  # Create the non root user and group.
  groupadd --gid ${GROUP_ID} ${GROUP_NAME} && \
  useradd --uid ${USER_ID} --gid ${GROUP_ID} --create-home ${USER_NAME} && \
  # Make the non root user the owner of the /go folder and their contents. This is required so the non root user can download and access the Golang dependency folder contents.
  chown -R -c ${USER_NAME} /go

USER ${USER_NAME}

# Development container stage, used for development locally from inside a container. See ./README.md for more details.
FROM base AS dev-container

ENV GIT_EDITOR="code --wait"

USER root

RUN \
  # Install tools.
  apt-get update && \
  # TODO: lock_versions to ensure deterministic behaviour.
  apt-get install -y \
  # Install zsh, used for a more modern shell in the development environment.
  zsh \
  # Install less, used to improve the shell read experience for bigger files.
  less \
  # Install sudo, used to allow the dev user to gain root level privileges.
  sudo && \
  # Clean up apt update and install unused artifacts.
  apt-get clean && \
  apt-get autoremove && \
  rm -rf /var/lib/apt/lists/* && \
  # Install sqlc, used to test the plugin locally.
  mkdir -p /tmp/sqlc && \
  curl -Ls https://github.com/sqlc-dev/sqlc/releases/download/v1.28.0/sqlc_1.28.0_linux_amd64.tar.gz -o /tmp/sqlc/sqlc.tar.gz && \
  tar xzf /tmp/sqlc/sqlc.tar.gz --directory /tmp/sqlc && \
  mv /tmp/sqlc/sqlc /usr/local/bin/sqlc && \
  rm -Rf /tmp/sqlc && \
  # Set zsh as the default shell for the root user.
  chsh -s $(which zsh) && \
  # Install Oh My Zsh (to improve the development shell experience) for the root user.
  sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
  # Set zsh as the default shell for the non root user.
  chsh -s $(which zsh) ${USER_NAME} && \
  # Allow the non root user to assume root privileges via sudo.
  adduser ${USER_NAME} sudo && \
  echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER ${USER_NAME}

# Install the VS Code Golang extension dependencies.
RUN go install github.com/cweill/gotests/gotests@v1.6.0 && \
  go install github.com/fatih/gomodifytags@v1.16.0 && \
  go install github.com/josharian/impl@v1.1.0 && \
  go install github.com/haya14busa/goplay/cmd/goplay@v1.0.0 && \
  go install github.com/go-delve/delve/cmd/dlv@v1.21.2 && \
  go install honnef.co/go/tools/cmd/staticcheck@v0.4.6 && \
  go install golang.org/x/tools/gopls@v0.14.2

# Install Oh My Zsh for the non root user.
RUN sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

CMD [ "bash", "-c", "echo 'Dev container started, sleeping' &&  while :; do sleep 1; done;" ]

# Dependency cache stage, used as a cache for Golang dependencies.
FROM base AS dependency-cache

COPY go.mod go.sum ./

RUN go mod download
#   go mod verify

# CI stage, used to run CI tasks.
FROM dependency-cache AS ci

COPY . .
