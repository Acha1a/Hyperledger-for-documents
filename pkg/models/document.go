package models

import "time"

type Document struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Hash      string `json:"hash"`
	Path      string `json:"path" db:"path"`          
	MimeType  string `json:"mime_type" db:"mime_type"` 
	FileSize  int64     `json:"file_size" db:"file_size"`
	Signature string `json:"signature"` 
	PublicKey string `json:"publicKey"`
	Owner     string `json:"owner"`
	Status    string `json:"status"`   
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Version   int    `json:"version"`
}