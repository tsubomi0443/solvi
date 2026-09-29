package repository

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	logutils "solvi/internal/shared/logUtils"
	"strings"

	"github.com/disintegration/imaging"
)

const opFileRepo = "FireRepository"

type FileRepository struct {
	uploadDir string
}

func NewFileRepository(uploadDir string) *FileRepository {
	return &FileRepository{
		uploadDir: uploadDir,
	}
}

func (repo *FileRepository) CreateFile(ctx context.Context, path string) (io.ReadWriteCloser, error) {
	const op = opFileRepo + ".CreateFile"
	f, err := os.Create(path)
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "アイコンファイル作成失敗", slog.String("path", path), slog.String("err", err.Error()))
		return nil, err
	}
	return f, nil
}

func (repo *FileRepository) ResizeImage(ctx context.Context, userID uint, reader io.Reader) (io.Reader, error) {
	const op = opFileRepo + ".ResizeImage"
	orgImg, _, err := image.Decode(reader)
	if err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "画像データ読込失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return nil, err
	}

	var (
		buf        = bytes.NewBuffer(nil)
		resizedImg = imaging.Resize(orgImg, 256, 0, imaging.Lanczos)
	)
	if err := png.Encode(buf, resizedImg); err != nil {
		logutils.Error(ctx, logutils.LayerRepository, op, "リサイズ済み画像データエンコード失敗", slog.Uint64("user_id", uint64(userID)), slog.String("err", err.Error()))
		return nil, err
	}

	resizedIcon := bytes.NewReader(buf.Bytes())
	return resizedIcon, nil
}

func (repo *FileRepository) DeleteFile(ctx context.Context, iconName string) error {
	const op = opFileRepo + ".DeleteFile"

	trimmed := strings.TrimSpace(iconName)
	if trimmed == "" {
		return nil
	}
	base := filepath.Base(trimmed)
	if base == "." || base == ".." || base != trimmed {
		err := fmt.Errorf("不正なアイコンファイル名です: %s", iconName)
		logutils.Error(ctx, logutils.LayerRepository, op, "パストラバーサル検知", slog.String("icon_name", iconName), slog.String("err", err.Error()))
		return err
	}

	cleanUploadDir := filepath.Clean(repo.uploadDir)
	targetPath := filepath.Clean(filepath.Join(cleanUploadDir, base))

	rel, err := filepath.Rel(cleanUploadDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		err := fmt.Errorf("無効なアイコンパスです: %s", iconName)
		logutils.Error(ctx, logutils.LayerRepository, op, "不正なパス", slog.String("icon_name", iconName), slog.String("err", err.Error()))
		return err
	}

	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "削除対象ファイルが存在しないためスキップ", slog.String("path", targetPath))
			return nil
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "アイコンファイル状態確認失敗", slog.String("path", targetPath), slog.String("err", err.Error()))
		return fmt.Errorf("アイコンファイル状態確認失敗: %w", err)
	}

	if !info.Mode().IsRegular() {
		err := fmt.Errorf("削除対象が通常ファイルではありません: %s (mode: %v)", targetPath, info.Mode())
		logutils.Error(ctx, logutils.LayerRepository, op, "通常ファイル外検出", slog.String("path", targetPath), slog.String("err", err.Error()))
		return err
	}

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			logutils.Debug(ctx, logutils.LayerRepository, op, "削除時にファイルが存在しなくなっていたためスキップ", slog.String("path", targetPath))
			return nil
		}
		logutils.Error(ctx, logutils.LayerRepository, op, "アイコンファイル削除失敗", slog.String("path", targetPath), slog.String("err", err.Error()))
		return fmt.Errorf("アイコンファイル削除失敗: %w", err)
	}

	logutils.Info(ctx, logutils.LayerRepository, op, "アイコンファイル削除完了", slog.String("path", targetPath))
	return nil
}
