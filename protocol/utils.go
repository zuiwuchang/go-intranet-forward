package protocol

import (
	"crypto/sha512"
	"encoding/hex"
)

// Hash .
func Hash(key, val string) (str string, e error) {
	val = key + val
	sha := sha512.New()
	_, e = sha.Write([]byte([]byte(val)))
	if e != nil {
		return
	}
	str = hex.EncodeToString(sha.Sum(nil))
	return
}
