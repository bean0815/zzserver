//包含 MD5 base64 urlcode Aes 等
package zztools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/url"
)

// Coding_ToMd5 加密
func Coding_ToMd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

//Coding_Base64Encode
func Coding_Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func Coding_Base64Decode(encodeString string) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(encodeString)
	if err != nil {
		return nil, err
	}
	return decodeBytes, nil
}

//Coding_UrlEncoding
func Coding_UrlEncoding(urltest string) string {
	return url.QueryEscape(urltest)
}
func Coding_UrlDecoder(encodeurl string) (string, error) {
	return url.QueryUnescape(encodeurl)
}

//Aes加密
func Coding_AesEncryptSimple(origData []byte, key string, iv string) ([]byte, error) {
	return Coding_AesDecryptPkcs5(origData, []byte(key), []byte(iv))
}

func Coding_AesEncryptPkcs5(origData []byte, key []byte, iv []byte) ([]byte, error) {
	return Coding_AesEncrypt(origData, key, iv, PKCS5Padding)
}

func Coding_AesEncrypt(origData []byte, key []byte, iv []byte, paddingFunc func([]byte, int) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = paddingFunc(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//Aes解密
func Coding_AesDecryptSimple(crypted []byte, key string, iv string) ([]byte, error) {
	return Coding_AesDecryptPkcs5(crypted, []byte(key), []byte(iv))
}

func Coding_AesDecryptPkcs5(crypted []byte, key []byte, iv []byte) ([]byte, error) {
	return Coding_AesDecrypt(crypted, key, iv, PKCS5UnPadding)
}

func Coding_AesDecrypt(crypted, key []byte, iv []byte, unPaddingFunc func([]byte) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = unPaddingFunc(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte("unpadding error")
	}
	return origData[:(length - unpadding)]
}
