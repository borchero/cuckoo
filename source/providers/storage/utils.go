package storage

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func getMimeType(filename string, file *os.File) (string, error) {
	// 1) Check if mime type get be determined from filename
	if strings.HasSuffix(filename, ".html") {
		return "text/html", nil
	} else if strings.HasSuffix(filename, ".css") {
		return "text/css", nil
	} else if strings.HasSuffix(filename, ".js") {
		return "application/javascript", nil
	}

	// 2) Otherwise, detect content type
	head := make([]byte, 512)
	if _, err := file.Read(head); err != nil {
		return "", fmt.Errorf("Failed to read head of file: %s", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("Failed to seek to beginning of file: %s", err)
	}
	return http.DetectContentType(head), nil
}
