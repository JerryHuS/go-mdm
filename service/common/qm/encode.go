/**
 * @Author: alessonhu
 * @Description:
 * @File:  encode.go
 * @Version: 1.0.0
 * @Date: 2021/3/10 19:28
 */
package qm

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

//  zip解压缩
func DoZlibUnCompress(data []byte) []byte {
	b := bytes.NewReader(data)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return []byte("")
	}
	io.Copy(&out, r)
	return out.Bytes()
}

// zip 压缩
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

//  md5加密
func Md5Encode(data []byte) string {
	h := md5.New()
	h.Write(data)
	//h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

//  AES CBC 模式解密
func AesCBCDecryptIv(input string, key []byte, iv []byte) ([]byte, error) {
	//先解密base64
	ciphertext, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("invalid decrypt key")
	}
	blockSize := block.BlockSize()
	if len(ciphertext) < blockSize {
		return nil, errors.New("ciphertext too short")
	}

	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plaintext, ciphertext)
	plaintext = PKCS5UnPaddingAes(plaintext)
	//plaintext = PKCS5UnPadding(plaintext)
	//PKCS5UnPaddingAes
	return plaintext, nil
}

func PKCS5UnPaddingAes(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	if unpadding >= length {
		return origData[:0]
	}
	return origData[:(length - unpadding)]
}
