package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/pkg/errors"
	ra "math/rand"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

//string->byte->hash
func GenBytesPrivateKey(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func AESEncrypt(key []byte, text []byte) (string, error) {
	//cipher.Block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err,"NewCipher error")
	}
	blockSize := block.BlockSize()
	originData := pad(text, blockSize)

	iv := []byte(GenCode(blockSize))
	blockMode := cipher.NewCBCEncrypter(block, iv)

	c := make([]byte, len(originData))
	blockMode.CryptBlocks(c, originData)

	var buffer bytes.Buffer
	buffer.Write(iv)
	buffer.Write(c)
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

//has base64
func AESDecrypt(key []byte, text string) (string, error) {
	decode_data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := decode_data[:16]
	blockMode := cipher.NewCBCDecrypter(block, iv)

	origin_data := make([]byte, len(decode_data[16:]))
	blockMode.CryptBlocks(origin_data, decode_data[16:])

	return string(unpad(origin_data)), nil
}

func unpad(ciphertext []byte) []byte {
	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

func RSAPKCS1V15Encrypt(publicKey *rsa.PublicKey, message []byte) (string, error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, message)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}
func RSAPKCS1V15Decrypt(privateKey *rsa.PrivateKey, base64string string) ([]byte, error) {
	decode_data, err := base64.StdEncoding.DecodeString(base64string)
	if err != nil {
		return nil, err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decode_data)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

type OAEP struct {
	hash   hash.Hash
	random io.Reader
}

// NewRSAOaep set attributes for an OAEP instance
//
// @param hash <Hash> hash function
// @return error
func NewRSAOaep(hash hash.Hash) *OAEP {
	oaep := &OAEP{}
	oaep.hash = hash
	oaep.random = rand.Reader

	return oaep
}

// Encrypt encrypts the given message with RSA-OAEP.
//
// @param publicKey <*rsa.PublicKey> public part of an RSA key
// @param message <[]byte> message to encrypt
// @param label <[]byte>
// @return ([]byte, error)
func (oaep *OAEP) RSAEncrypt(publicKey *rsa.PublicKey, message []byte, label []byte) (string, error) {
	cipherText, err := rsa.EncryptOAEP(oaep.hash, oaep.random, publicKey, message, label)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Dencrypt dencrypts the given message with RSA-OAEP.
//
// @param privateKey <*rsa.PrivateKey> represents an RSA key
// @param label <[]byte> label parameter must match the value given when encrypting
// @return ([]byte, error)
func (oaep *OAEP) RSADencrypt(privateKey *rsa.PrivateKey, base64string string, label []byte) ([]byte, error) {
	decode_data, err := base64.StdEncoding.DecodeString(base64string)
	if err != nil {
		return nil, err
	}
	plaintext, err := rsa.DecryptOAEP(oaep.hash, oaep.random, privateKey, decode_data, label)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

/**
  * @parameter: file 公钥文件路径
  * @return: 公钥 ;错误信息
  * @Description: 将对应路径下的公钥解析出RSA公钥
  * @author: shalom
  * @date: 2022/1/27 11:37 上午
  * @version: V1.0
  */
func ParseRSAPubKey(file string) (*rsa.PublicKey, error) {
	publicKey, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err,"ReadFile error")
	}
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("block is nil")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err,"x509.ParsePKIXPublicKey error")
	}
	return pubInterface.(*rsa.PublicKey), nil
}

func ParsePrivateKey(file string) (*rsa.PrivateKey, error) {
	prv, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(prv)
	if block == nil {
		return nil, errors.New("block is nil")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

//use for random code
func GenCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	x := len(numeric)
	ra.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[ra.Intn(x)])
	}
	return sb.String()
}