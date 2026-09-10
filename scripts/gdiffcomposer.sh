#!/bin/bash
# 概要
#   指定した2つのコミット間の差分を取得し、
#   gdiff.shで生成した複数の diff_*.md を
#   combined_patches.patch.md に結合します。
#
# 使用方法
#   ./combine_patches.sh [-o OUTPUT_DIR] <BASE_COMMIT> <TARGET_COMMIT>
#
# オプション
#   -o OUTPUT_DIR
#       出力先ディレクトリ（デフォルト: .）
#
# 例
#   ./combine_patches.sh 1234567 abcdef0
#   ./combine_patches.sh -o ./output 1234567 abcdef0
#
# 出力
#   diff_*.md
#   combined_patches.patch.md
#
# 注意
#   - gdiff.sh が同じディレクトリに存在する必要があります。
#   - 既存の combined_patches.patch.md は上書きされます。

# 1. デフォルト値の定義
OUTPUT_DIR="."
COMBINED_FILE="combined_patches.patch.md"

# 2. 自身のスクリプトがあるディレクトリを取得 (gdiff.sh を同じ場所に置いている場合を想定)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 3. フラグの解析
while getopts "o:" opt; do
  case $opt in
    o) OUTPUT_DIR="$OPTARG" ;;
    *) echo "Usage: $0 [-o output_dir] <base_commit> <target_commit>"; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

# 4. 位置引数（コミットハッシュ）のチェック
if [ "$#" -lt 2 ]; then
    echo "Usage: $0 [-o output_dir] <base_commit> <target_commit>"
    exit 1
fi

BASE_COMMIT=$1
TARGET_COMMIT=$2

# 5. 作成済みの gdiff.sh を呼び出して個別パッチを生成
echo "=== 1. Calling gdiff.sh ==="
# -o フラグを引き継いで gdiff.sh を実行
bash "${SCRIPT_DIR}/gdiff.sh" -o "$OUTPUT_DIR" "$BASE_COMMIT" "$TARGET_COMMIT"

if [ $? -ne 0 ]; then
    echo "[ERROR] gdiff.sh failed. Exiting."
    exit 1
fi

# 6. 結合処理の実行
FINAL_OUTPUT_PATH="${OUTPUT_DIR}/${COMBINED_FILE}"
# 結合ファイルを初期化
> "$FINAL_OUTPUT_PATH"

echo ""
echo "=== 2. Combining Patches ==="

# 生成された個別パッチファイルをソートしてループ処理
PATCH_FILES=$(ls "${OUTPUT_DIR}"/diff_*.md 2>/dev/null | sort)

if [ -z "$PATCH_FILES" ]; then
    echo "[ERROR] Generated patch files not found in $OUTPUT_DIR"
    exit 1
fi

for FILE_PATH in $PATCH_FILES; do
    # フォルダ名を除いたファイル名のみを取得 (例: diff_20260520_01.patch.md)
    BASE_FILE_NAME=$(basename "$FILE_PATH")
    
    echo "Inlining: ${BASE_FILE_NAME}"
    
    # 指定されたフォーマットで結合ファイルへ追記
    echo "-----" >> "$FINAL_OUTPUT_PATH"
    echo "{${BASE_FILE_NAME}}" >> "$FINAL_OUTPUT_PATH"
    cat "$FILE_PATH" >> "$FINAL_OUTPUT_PATH"
    echo "----------" >> "$FINAL_OUTPUT_PATH"
done

echo "--------------------------------------------------"
echo "[SUCCESS] Combined file generated at: $FINAL_OUTPUT_PATH"