package service

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
)

func SaveFile(file multipart.File, filename string) (string, string, error) {
	path := "storage/uploads/" + filename

	out, err := os.Create(path)
	if err != nil {
		return "", "", err
	}
	defer out.Close()

	hash := sha256.New()
	mw := io.MultiWriter(out, hash)

	_, err = io.Copy(mw, file)
	if err != nil {
		return "", "", err
	}

	return path, hex.EncodeToString(hash.Sum(nil)), nil
}