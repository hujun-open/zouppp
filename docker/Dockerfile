FROM ubuntu:24.04

ARG GO_VERSION=1.24.0
ARG DEBIAN_FRONTEND=noninteractive

# Mirror the CI "install dep" step (plus build/curl/sudo helpers)
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl git sudo \
        build-essential pkg-config \
        iproute2 iputils-ping net-tools tcpdump \
        libpcap-dev ppp pppoe kea-dhcp6-server \
    && rm -rf /var/lib/apt/lists/*

# Go 1.24 (matches actions/setup-go go-version)
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o /tmp/go.tgz \
    && tar -C /usr/local -xzf /tmp/go.tgz && rm /tmp/go.tgz
ENV PATH=/usr/local/go/bin:/root/go/bin:$PATH \
    GOPATH=/root/go \
    GOTOOLCHAIN=local

WORKDIR /src
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
