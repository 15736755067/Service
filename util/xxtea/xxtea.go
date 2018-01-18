package xxtea

import "github.com/xxtea/xxtea-go/xxtea"

var key = "e4e0c909c39dbee1f55507b1e2ffdfa3"

func XxteaEncrypt(data []byte) []byte {
	encryptData := xxtea.Encrypt(data, []byte(key))
	return encryptData
}

func XxteaDecrypt(data []byte) []byte {
	decryptData := xxtea.Decrypt(data, []byte(key))
	return decryptData
}
