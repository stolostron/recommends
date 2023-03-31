package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateID(object interface{}) string {

	objectString := fmt.Sprintf("%v", object)
	hash := sha256.Sum256([]byte(objectString))

	hexString := hex.EncodeToString(hash[:])

	return hexString
}
