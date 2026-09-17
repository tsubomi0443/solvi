package repository

import "io"

type FileRepository interface {
	CreateFile(writer io.Writer) error
	DeleteFile(name string) error
}
