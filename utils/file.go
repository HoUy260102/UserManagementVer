package utils

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

var validFile = map[string]bool{
	".png": true,
	".jpg": true,
}

var validMiMe = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

func ChechValidFile(fileHeader *multipart.FileHeader) error {
	fileName := fileHeader.Filename
	fileExt := filepath.Ext(fileName)
	if _, ok := validFile[fileExt]; !ok {
		return fmt.Errorf("%s is not a valid file", fileName)
	}
	return nil
}

func CheckValidMiMe(fileHeader *multipart.FileHeader) error {
	f, err := fileHeader.Open()
	
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)

	mimeType := http.DetectContentType(buf[:n])
	if _, ok := validMiMe[mimeType]; !ok {
		return fmt.Errorf("%s is not a valid mime type", mimeType)
	}

	return nil
}
