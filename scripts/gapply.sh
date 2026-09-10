#!/bin/bash
# 概要
#   diff_*.md または diff_*.patch.md を順番に適用します。
#   オプション指定時は、適用後にコミットも実行します。
#
# 使用方法
#   ./apply_patches.sh [-c|--is-commit]
#
# オプション
#   -c, --is-commit, -is-commit
#       パッチ適用後にコミットを実行します。
#
# 例
#   ./apply_patches.sh
#   ./apply_patches.sh -c
#
# 対象ファイル
#   diff_*.md
#   diff_*.patch.md
#
# 注意
#   - パッチはファイル名順に適用されます。
#   - 適用できない場合は 3-way マージを試行します。
#   - 3-way マージも失敗した場合は処理を終了します。
#   - コミット時は各パッチの1行目をコミットメッセージとして使用します。
#   - オプション未指定時はコミットを行いません。

# デフォルト設定 (コミットはしない)
IS_COMMIT=false

# 引数の解析
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -is-commit|--is-commit|-c) IS_COMMIT=true; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
done

echo "=== Starting Patch Application Process ==="
echo "Mode: [is-commit = ${IS_COMMIT}]"

# カレントディレクトリ内の該当パッチファイルを名前順でループ（diff_*.md / diff_*.patch.md）
for f in $(ls diff_*.md diff_*.patch.md 2>/dev/null | sort -u); do
    echo "--------------------------------------------------"
    echo "Processing: $f"

    # 1行目からコミットメッセージを取得
    COMMIT_MSG=$(head -n 1 "$f")
    
    # 3行目以降（パッチ本体）を切り出し
    PATCH_BODY=$(tail -n +3 "$f")

    # 【コツ1】 `--check` による事前の擬似実行 (ズレ無視も指定)
    echo "$PATCH_BODY" | git apply --check --ignore-whitespace --whitespace=nowarn - 2>/dev/null
    
    if [ $? -eq 0 ]; then
        # 擬似実行成功：通常適用
        echo "-> Check passed. Applying cleanly..."
        echo "$PATCH_BODY" | git apply --ignore-whitespace --whitespace=fix -
    else
        # 擬似実行失敗：【コツ3】 `--3way` マージを有効化
        echo "-> Check failed. Attempting 3-way merge..."
        echo "$PATCH_BODY" | git apply --ignore-whitespace --whitespace=fix --3way -
        
        if [ $? -ne 0 ]; then
            echo "[ERROR] 3-way merge failed for $f."
            echo "Please resolve conflicts manually, then restart the script."
            exit 1
        fi
        echo "-> 3-way merge applied (Conflicts may need review)."
    fi

    # フラグが有効な場合のみコミットを実行
    if [ "$IS_COMMIT" = true ]; then
        git add .
        if ! git diff-index --quiet HEAD --; then
            git commit -m "$COMMIT_MSG"
            echo "[SUCCESS] Applied and committed: \"$COMMIT_MSG\""
        else
            echo "[INFO] No changes to commit."
        fi
    else
        echo "[SUCCESS] Applied (Skipped commit as per flag)."
    fi

done

echo "--------------------------------------------------"
echo "=== All patches processed successfully! ==="