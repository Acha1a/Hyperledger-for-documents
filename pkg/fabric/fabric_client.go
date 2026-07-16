package fabric

import "log"

func CreateDocumentHash(hash string, owner string) error {
	log.Println("Writing hash to blockchain:", hash)

	return nil
}

func GetDocumentHash(hash string) (string, error) {

	log.Println("Checking hash in blockchain:", hash)

	return hash, nil
}