package entity

import (
	"io"
	"time"
)

type LogEntity struct {
	Name    string
	Content io.ReadCloser
}

type DownloadKind int

const (
	DownloadKindAll  DownloadKind = 1
	DownloadKindDate DownloadKind = 2
)

type DownloadTicket struct {
	Key       string
	Password  string
	UserUUID  string
	Kind      DownloadKind
	Date      time.Time
	ExpiresAt time.Time
	Filename  string
}
