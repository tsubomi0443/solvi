package repository

import (
	"context"
	"io"
)

type FileRepository interface {
	CreateFile(ctx context.Context, path string) (io.ReadWriteCloser, error)
	ResizeImage(ctx context.Context, userID uint, reader io.Reader) (io.Reader, error)
	DeleteFile(ctx context.Context, iconName string) error
}
