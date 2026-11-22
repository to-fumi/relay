#!/usr/bin/env bash
# Dev Container から外への通信を、GitHub と /etc/relay/allowed-domains.txt のドメインだけにする。
# コンテナを起動するたびに root で実行する（postStartCommand）。
# Anthropic の参考実装（anthropics/claude-code の .devcontainer/init-firewall.sh）をもとにしている。
# Claude Code の devcontainer feature が /usr/local/bin/init-firewall.sh に自分のものを置くので、名前を分けている。
set -euo pipefail
IFS=$'\n\t'

DOMAINS_FILE=/etc/relay/allowed-domains.txt

# 通信を絞る filter の表だけを作り直す。nat の表には Docker の内部 DNS（127.0.0.11）の設定があるので触らない
iptables -P INPUT ACCEPT
iptables -P OUTPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -F
iptables -X
ipset destroy allowed-domains 2>/dev/null || true

# DNS と localhost（db への接続を含む）は先に許可する
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
iptables -A INPUT -p udp --sport 53 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 53 -j ACCEPT
iptables -A INPUT -p tcp --sport 53 -j ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

ipset create allowed-domains hash:net

# ドメインの IP を引いて許可する。引けなかったものは飛ばして、警告だけ出す
allow_domain() {
  local domain=$1 ips ip
  ips=$(dig +noall +answer A "$domain" | awk '$4 == "A" {print $5}')
  if [ -z "$ips" ]; then
    echo "WARN: $domain の IP を引けなかったので飛ばす" >&2
    return 0
  fi
  while read -r ip; do
    if [[ "$ip" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
      ipset add -exist allowed-domains "$ip"
    fi
  done < <(echo "$ips")
  return 0
}

# GitHub（git、API、Web）の IP の範囲。API は認証なしだと回数制限があるので、3 回までやり直し、
# それでも取れなければ主なドメインの IP で代わりにする
gh_ranges=""
for attempt in 1 2 3; do
  gh_ranges=$(curl -s --connect-timeout 10 https://api.github.com/meta || true)
  if echo "$gh_ranges" | jq -e '.web and .api and .git' >/dev/null 2>&1; then
    break
  fi
  gh_ranges=""
  sleep $((attempt * 2))
done
if [ -n "$gh_ranges" ]; then
  while read -r cidr; do
    if [[ ! "$cidr" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}$ ]]; then
      echo "ERROR: GitHub の IP の範囲の形がおかしい: $cidr" >&2
      exit 1
    fi
    ipset add -exist allowed-domains "$cidr"
  done < <(echo "$gh_ranges" | jq -r '(.web + .api + .git)[]' | grep -v ':' | aggregate -q)
else
  echo "WARN: GitHub の IP の範囲を取れなかった（API の回数制限など）。主なドメインの IP で代わりにする" >&2
  for domain in github.com api.github.com codeload.github.com uploads.github.com \
    objects.githubusercontent.com raw.githubusercontent.com release-assets.githubusercontent.com; do
    allow_domain "$domain"
  done
fi

# 許可するドメイン（allowed-domains.txt）
while read -r domain; do
  [[ -z "$domain" || "$domain" == \#* ]] && continue
  allow_domain "$domain"
done < "$DOMAINS_FILE"

# ホスト（Mac）側のネットワークとはやり取りできるようにする（VS Code のポート転送など）
HOST_IP=$(ip route | awk '/default/ {print $3; exit}')
if [ -z "$HOST_IP" ]; then
  echo "ERROR: ホストの IP を取れなかった" >&2
  exit 1
fi
HOST_NETWORK=$(echo "$HOST_IP" | sed "s/\.[0-9]*$/.0\/24/")
iptables -A INPUT -s "$HOST_NETWORK" -j ACCEPT
iptables -A OUTPUT -d "$HOST_NETWORK" -j ACCEPT

iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT DROP
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m set --match-set allowed-domains dst -j ACCEPT
# それ以外はすぐに断る（待たされないように）
iptables -A OUTPUT -j REJECT --reject-with icmp-admin-prohibited

# 確かめる：許可していないところには届かず、GitHub と npm には届く
if curl -s --connect-timeout 5 https://example.com >/dev/null 2>&1; then
  echo "ERROR: 許可していない example.com に届いてしまった" >&2
  exit 1
fi
# GitHub は回数制限のない github.com で確かめる
for url in https://github.com/ https://registry.npmjs.org/; do
  if ! curl -s --connect-timeout 5 "$url" >/dev/null 2>&1; then
    echo "ERROR: $url に届かない" >&2
    exit 1
  fi
done
echo "ファイアウォールを設定した（許可：GitHub と $DOMAINS_FILE のドメイン）"
