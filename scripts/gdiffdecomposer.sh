#!/bin/bash
# 概要
#   combined_patches.patch.md を解析し、
#   結合されたパッチを個別の diff_*.md ファイルへ復元します。
#
# 使用方法
#   ./split_patches.sh [-o OUTPUT_DIR] [INPUT_FILE]
#
# オプション
#   -o OUTPUT_DIR
#       出力先ディレクトリ（デフォルト: .）
#
# 引数
#   INPUT_FILE
#       分解対象の統合パッチファイル
#       （デフォルト: combined_patches.patch.md）
#
# 例
#   ./split_patches.sh
#   ./split_patches.sh combined_patches.patch.md
#   ./split_patches.sh -o ./output combined_patches.patch.md
#
# 出力
#   diff_YYYYMMDD_01.patch.md
#   diff_YYYYMMDD_02.patch.md
#   ...
#
# 注意
#   - 出力ディレクトリが存在する場合は削除して再作成します。
#   - {diff_xxx.md} の情報をもとにファイルを復元します。
#   - 入力ファイルが存在しない場合はエラー終了します。

# デフォルト値
INPUT_FILE="combined_patches.patch.md"
OUTPUT_DIR="."

# フラグの解析
while getopts "o:" opt; do
  case $opt in
    o) OUTPUT_DIR="$OPTARG" ;;
    *) echo "Usage: $0 [-o output_dir] [input_file]"; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

# 位置引数で入力ファイル名が指定されていれば上書き
if [ "$#" -gt 0 ]; then
    INPUT_FILE="$1"
fi

# 入力ファイルの存在チェック
if [ ! -f "$INPUT_FILE" ]; then
    echo "[ERROR] Input file not found: $INPUT_FILE"
    exit 1
fi

# 出力先ディレクトリの初期化（既存なら削除して新規作成）
if [ "$OUTPUT_DIR" != "." ]; then
    if [ -d "$OUTPUT_DIR" ]; then
        echo "Clearing existing directory: $OUTPUT_DIR"
        rm -rf "$OUTPUT_DIR"
    fi
    mkdir -p "$OUTPUT_DIR"
fi

echo "=== Starting Patch Decomposition ==="
echo "Input file: $INPUT_FILE"
echo "Output dir: $OUTPUT_DIR"

# 状態管理フラグと変数
IN_PATCH=false
CURRENT_FILE=""
TMP_CONTENT=""

# 統合ファイルを1行ずつ読み込み処理
while IFS= read -r line || [ -n "$line" ]; do
    # 開始マーカーの検出
    if [ "$line" = "-----" ] && [ "$IN_PATCH" = false ]; then
        IN_PATCH=true
        TMP_CONTENT=""
        continue
    fi
    
    # 終了マーカーの検出
    if [ "$line" = "----------" ] && [ "$IN_PATCH" = true ]; then
        if [ -n "$CURRENT_FILE" ]; then
            # 最後の不要な改行を1つ削ってファイルに書き出し
            echo -n "$TMP_CONTENT" | sed '$d' > "${OUTPUT_DIR}/${CURRENT_FILE}"
            echo "Extracted: ${CURRENT_FILE}"
        fi
        IN_PATCH=false
        CURRENT_FILE=""
        continue
    fi

    # パッチ内部の処理
    if [ "$IN_PATCH" = true ]; then
        # 1行目がファイル名フォーマット {diff_...} かどうかの判定
        if [ -z "$CURRENT_FILE" ] && [[ "$line" =~ ^\{diff_.*\.md\}$ ]]; then
            # 波括弧を取り除いてファイル名を抽出
            CURRENT_FILE=$(echo "$line" | sed 's/[{}]//g')
        else
            # パッチ本文（コミットメッセージ含む）を蓄積
            TMP_CONTENT+="${line}"$'\n'
        fi
    fi
done < "$INPUT_FILE"

echo "--------------------------------------------------"
echo "[SUCCESS] All patches decomposed into: $OUTPUT_DIR"