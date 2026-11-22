#!/usr/bin/env bash
# Dev Container を作ったあとに 1 回だけ実行する
set -euo pipefail

# db とネットワークを共有しているので、作り直す前のコンテナのファイアウォール（古い許可リスト）が残っていることがある。
# セットアップの間は外し、終わったら postStartCommand で最新の許可リストでかけ直す
sudo iptables -P INPUT ACCEPT
sudo iptables -P OUTPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -F

# volume は root の持ち物で作られるので、vscode ユーザーのものにする
sudo chown -R vscode:vscode /home/vscode/.claude

# pnpm は package.json の packageManager で決めたバージョンを使う
corepack enable
corepack install

pnpm install --frozen-lockfile

# Go の toolchain と依存を先に取ってくる（初回の go run を速くする）
(cd backend-go && go mod download && go mod download -modfile=tools/go.mod)

# E2E で使う Chromium と、動かすのに要るパッケージ
pnpm task e2e:install

# origin が SSH（git@github.com:）でも、コンテナの中では HTTPS と gh のトークンで通信する（SSH の鍵は渡していないため）
git config --global url."https://github.com/".insteadOf "git@github.com:"

# 開発用の設定（なければ作る。仮ログインで動く）
[ -f backend-go/.env ] || cp backend-go/.env.example backend-go/.env
[ -f web/.env ] || cp web/.env.example web/.env

# ここから先、vscode ユーザーの sudo はファイアウォールの再設定だけにする
# （AI がファイアウォールを外したり、パッケージを入れたりできないように。要るときはコンテナを作り直す）
echo 'vscode ALL=(root) NOPASSWD: /usr/local/bin/relay-firewall.sh' | sudo tee /etc/sudoers.d/relay-firewall >/dev/null
sudo chmod 0440 /etc/sudoers.d/relay-firewall
sudo rm -f /etc/sudoers.d/vscode
if sudo -n true 2>/dev/null; then
  echo "ERROR: sudo をまだ何にでも使える" >&2
  exit 1
fi

echo "Dev Container の準備ができた。pnpm task dev:api と pnpm task dev:web で起動する"
