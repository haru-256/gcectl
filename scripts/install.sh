#!/bin/sh

set -e # エラーが発生したらスクリプトを停止

# --- 変数を設定 (リポジトリに合わせて変更) ---
REPO_OWNER="haru-256"
REPO_NAME="gcectl"
BINARY_NAME="gcectl"
INSTALL_DIR="${HOME}/.local/bin"

# --- 1. OSとアーキテクチャを判別 ---
OS_TYPE=$(uname -s)
ARCH_TYPE=$(uname -m)

case $OS_TYPE in
  Linux)
    OS_NAME="linux"
    ;;
  Darwin) # macOS
    OS_NAME="darwin"
    ;;
  *)
    echo "サポートされていないOSです: $OS_TYPE"
    exit 1
    ;;
esac

case $ARCH_TYPE in
  x86_64 | amd64)
    ARCH_NAME="amd64"
    ;;
  arm64 | aarch64)
    ARCH_NAME="arm64"
    ;;
  *)
    echo "サポートされていないアーキテクチャです: $ARCH_TYPE"
    exit 1
    ;;
esac

# --- 2. 最新リリースのバージョンとダウンロードURLを取得 ---
echo "最新バージョンを確認中..."

# GitHub APIから最新リリース情報を取得
RELEASE_JSON=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest")

# バージョンを抽出（"v0.0.1" の形式）
VERSION=$(echo "$RELEASE_JSON" | grep -o '"tag_name": *"[^"]*"' | cut -d '"' -f 4)

if [ -z "$VERSION" ]; then
  echo "エラー: 最新バージョンの取得に失敗しました。"
  exit 1
fi

# "v" プレフィックスを除去（v0.0.1 → 0.0.1）
VERSION_NUMBER=$(echo "$VERSION" | sed 's/^v//')

echo "最新バージョン: $VERSION"

# --- 3. GoReleaserの命名規則に合わせてファイル名を生成 ---
# 形式: gcectl_{version}_{os}_{arch}.tar.gz
ARCHIVE_NAME="${REPO_NAME}_${VERSION_NUMBER}_${OS_NAME}_${ARCH_NAME}.tar.gz"

echo "ダウンロード対象: $ARCHIVE_NAME"

# --- 4. ダウンロードURLを取得 ---
DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep -o "\"browser_download_url\": *\"https://[^\"]*${ARCHIVE_NAME}\"" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "エラー: ${ARCHIVE_NAME} のダウンロードURLが見つかりません。"
  echo "利用可能なファイル:"
  echo "$RELEASE_JSON" | grep "browser_download_url" | cut -d '"' -f 4
  exit 1
fi

echo "${ARCHIVE_NAME} をダウンロードしています..."

# --- 5. ダウンロードとインストール ---
TEMP_DIR=$(mktemp -d); trap 'rm -rf "$TEMP_DIR"' EXIT
curl -L "$DOWNLOAD_URL" -o "${TEMP_DIR}/${ARCHIVE_NAME}"

echo "インストール中..."
tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"

# インストールディレクトリが存在しない場合は作成
mkdir -p "$INSTALL_DIR"

# install コマンド (または mv) を使ってパスが通った場所に配置
# sudoが必要な場合がある
if [ -w "$INSTALL_DIR" ]; then
  install -m 755 "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
else
  echo "管理者権限 (sudo) が必要です: ${INSTALL_DIR} に ${BINARY_NAME} を移動します"
  sudo install -m 755 "${TEMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
fi

echo ""
echo "✅ ${BINARY_NAME} ${VERSION} のインストールが完了しました。"
echo ""
echo "インストール先: ${INSTALL_DIR}/${BINARY_NAME}"
echo ""

# PATHが通っているか確認
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo "⚠️  注意: ${INSTALL_DIR} が PATH に含まれていません。"
  echo ""
  echo "以下のコマンドを実行して PATH に追加してください:"
  echo ""
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
  echo "  source ~/.bashrc"
  echo ""
  echo "（zshの場合は ~/.zshrc を使用）"
  echo ""
fi

echo "バージョン確認:"
if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
  ${BINARY_NAME} --version
else
  ${INSTALL_DIR}/${BINARY_NAME} --version
fi
