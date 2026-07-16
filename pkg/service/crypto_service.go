package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

var MinisterKey *rsa.PrivateKey

func GenerateMinisterKeys() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func SignHash(docHash string, privKey *rsa.PrivateKey) (string, error) {
	hashBytes, err := hex.DecodeString(docHash)
	if err != nil {
		return "", fmt.Errorf("invalid hash format: %v", err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %v", err)
	}

	return hex.EncodeToString(signature), nil
}

func ExportPublicKey(privKey *rsa.PrivateKey) string {
	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	return string(pubKeyPem)
}