package utils

import "golang.org/x/crypto/bcrypt"

func HashID(id string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(id), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckID(id, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(id))
	return err == nil
}
