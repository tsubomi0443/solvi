#!/bin/bash
# 概要
#   指定した2つのコミット間の差分をコミット単位で取得し、
#   diff_YYYYMMDD_XX.patch.md 形式のパッチファイルを生成します。
#
# 使用方法
#   ./gdiff.sh [-o OUTPUT_DIR] <BASE_COMMIT> <TARGET_COMMIT>
#
# オプション
#   -o OUTPUT_DIR
#       パッチファイルの出力先ディレクトリ（デフォルト: .）
#
# 例
#   ./gdiff.sh 1234567 abcdef0
#   ./gdiff.sh -o ./output 1234567 abcdef0
#
# 出力
#   diff_YYYYMMDD_01.patch.md
#   diff_YYYYMMDD_02.patch.md
#   ...
#
# 注意
#   - 各ファイルの先頭にコミットメッセージが出力されます。
#   - パッチはコミットの古い順に生成されます。
#   - 出力ディレクトリが存在する場合は削除して再作成します。
#   - diff_*.md ファイルは差分対象から除外されます。

# デフォルト値
OUTPUT_DIR="."

# フラグの解析
while getopts "o:" opt; do
  case $opt in
    o) OUTPUT_DIR="$OPTARG" ;;
    *) echo "Usage: $0 [-o output_dir] <base_commit> <target_commit>"; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 [-o output_dir] <base_commit> <target_commit>"
    exit 1
fi

BASE_COMMIT=$1
TARGET_COMMIT=$2

if [ "$OUTPUT_DIR" != "." ]; then
    if [ -d "$OUTPUT_DIR" ]; then
        echo "Clearing existing directory: $OUTPUT_DIR"
        rm -rf "$OUTPUT_DIR"
    fi
    mkdir -p "$OUTPUT_DIR"
fi

COMMITS=$(git log --reverse --format=%H "${BASE_COMMIT}..${TARGET_COMMIT}")
CURRENT_BASE=$BASE_COMMIT
INDEX=1

for COMMIT in $COMMITS; do
    COMMIT_DATE=$(git show -s --format=%cd --date=format:'%Y%m%d' "$COMMIT")
    # コミットメッセージを取得 (1行目のみ: %s)
    COMMIT_MSG=$(git show -s --format='%s' "$COMMIT")
    PREFIX=$(printf "%02d" $INDEX)
    
    FILE_NAME="${OUTPUT_DIR}/diff_${COMMIT_DATE}_${PREFIX}.patch.md"
    echo "Generating: ${FILE_NAME}"
    
    # 1行目: コミットメッセージ
    echo "$COMMIT_MSG" > "${FILE_NAME}"
    # 2行目: 空行
    echo "" >> "${FILE_NAME}"
    # 3行目〜: パッチ本体
    git diff -U50 "${CURRENT_BASE}" "${COMMIT}" -- . ":(exclude)diff_*.md" >> "${FILE_NAME}"
    
    CURRENT_BASE=$COMMIT
    INDEX=$((INDEX + 1))
done

echo "All patches generated in: $OUTPUT_DIR"
