# Dev Container の土台。外への通信を許可したドメインだけに絞るための道具と、その設定を入れる
FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    iptables ipset dnsutils jq aggregate iproute2 \
  && rm -rf /var/lib/apt/lists/*

# 許可するドメインはイメージに焼き込む。コンテナの中から書き換えて広げられないようにするため
# （変えるときは allowed-domains.txt を PR で直し、コンテナを作り直す）
COPY allowed-domains.txt /etc/relay/allowed-domains.txt
COPY relay-firewall.sh /usr/local/bin/relay-firewall.sh
RUN chmod 644 /etc/relay/allowed-domains.txt && chmod 755 /usr/local/bin/relay-firewall.sh
