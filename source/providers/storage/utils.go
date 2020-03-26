package storage

import (
	"fmt"
	"net/http"
	"os"
)

func getMimeType(file *os.File) (string, error) {
	head := make([]byte, 512)
	if _, err := file.Read(head); err != nil {
		return "", fmt.Errorf("Failed to read head of file: %s", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("Failed to seek to beginning of file: %s", err)
	}
	return http.DetectContentType(head), nil
}
