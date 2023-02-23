package generate

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetKey(date string) string {
	prefix := "enedge"

	source := prefix[0:2] + date[4:6] + date[0:4] + prefix[2:]

	hash := sha256.Sum256([]byte(source))

	return hex.EncodeToString(hash[:])
}
