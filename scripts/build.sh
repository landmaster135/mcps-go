#!/bin/bash
function build_go_bin(){
  local output_dir=$1
  local output_name=$2
  local package=$3

  # プラットフォーム別の出力先
  LINUX_AMD64_DIR="${output_dir}/linux_amd64"
  MAC_ARM64_DIR="${output_dir}/darwin_arm64"
  WIN_AMD64_DIR="${output_dir}/win_amd64"
  WIN_OUTPUT_NAME="${output_name}.exe"

  echo "Building ${output_name}..."

  # Linux/AMD64向けビルド
  echo "Building for Linux/AMD64..."
  mkdir -p "${LINUX_AMD64_DIR}"
  GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "${LINUX_AMD64_DIR}/${output_name}" "${package}"

  # Windows/AMD64向けビルド
  echo "Building for Windows/AMD64..."
  mkdir -p "${WIN_AMD64_DIR}"
  GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "${WIN_AMD64_DIR}/${WIN_OUTPUT_NAME}" "${package}"

  # macOS/ARM64向けビルド
  echo "Building for macOS/ARM64..."
  mkdir -p "${MAC_ARM64_DIR}"
  GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o "${MAC_ARM64_DIR}/${output_name}" "${package}"

  echo "Build completed successfully!"
}

# エラーが発生したら終了
set -e

# ビルド対象のディレクトリ
CMD_DIR="."

# 出力先ディレクトリ
OUTPUT_DIR="./pkg/bin"
# ビルド情報
OUTPUT_NAME="mcps-go"
PACKAGE="github.com/landmaster135/mcps-go/${CMD_DIR}"

build_go_bin $OUTPUT_DIR $OUTPUT_NAME $PACKAGE
