package repository

import "io"

type FileRepository struct {
}

func (repo *FileRepository) CreateFile(writer io.Writer) error {
	return nil
}

func (repo *FileRepository) DeleteFile(name string) error {
	return nil
}
