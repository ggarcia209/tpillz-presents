package hashops

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// Get shortened version of hash
func GetMD5Hash64Bit(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:4])
}
