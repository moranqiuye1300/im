package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
)

func Md5Encode(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))
	temp := hash.Sum(nil)
	return hex.EncodeToString(temp)
}
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}
func MakePassword(plainpwd, salt string) string {
	return MD5Encode(plainpwd + salt)
}
func ValidatePassword(plainpwd, salt, hash string) bool {
	return MakePassword(plainpwd, salt) == hash
}

// RandomSalt 生成随机盐
func RandomSalt(length int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}
