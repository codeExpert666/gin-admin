// Package aes 提供AES加密解密功能
// 本包实现了基于AES-CBC模式的加密解密操作，并提供了Base64编码的便捷方法
package aes

import (
	"bytes"      // 用于处理字节切片
	"crypto/aes" // Go标准库中的AES加密包
	"crypto/cipher" // 提供了加密算法的块模式
	"encoding/base64" // 用于Base64编码解码
)

var (
	// SecretKey 是AES加密的密钥
	// 这里使用32字节(256位)的密钥，可以用于AES-256加密
	// 在实际应用中，建议通过配置文件或环境变量来设置密钥
	SecretKey = []byte("2985BCFDB5FE43129843DB59825F8647")
)

// PKCS5Padding 实现PKCS#5填充
// 在AES加密时，明文必须是块大小的整数倍，如果不是则需要进行填充
// 参数:
//   - plaintext: 需要填充的原始数据
//   - blockSize: 块大小(AES固定为16字节)
// 返回:
//   - 填充后的数据
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	// 计算需要填充的长度
	padding := blockSize - len(plaintext)%blockSize
	// 创建填充字节切片，填充值为填充的长度
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	// 将填充添加到原始数据末尾
	return append(plaintext, padtext...)
}

// PKCS5UnPadding 实现PKCS#5去填充
// 在解密后，需要删除之前添加的填充数据
// 参数:
//   - origData: 解密后的数据（包含填充）
// 返回:
//   - 去除填充后的原始数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 获取填充的长度，填充值的最后一个字节即为填充长度
	unpadding := int(origData[length-1])
	// 返回去除填充后的数据
	return origData[:(length - unpadding)]
}

// Encrypt 使用AES-CBC模式加密数据
// 参数:
//   - origData: 需要加密的原始数据
//   - key: 加密密钥，必须是16、24或32字节（对应AES-128、AES-192、AES-256）
// 返回:
//   - 加密后的数据和可能的错误
func Encrypt(origData, key []byte) ([]byte, error) {
	// 创建AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 获取块大小（AES固定为16字节）
	blockSize := block.BlockSize()
	// 对数据进行PKCS#5填充
	origData = PKCS5Padding(origData, blockSize)
	// 使用CBC模式加密，IV（初始化向量）使用密钥的前blockSize个字节
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 创建保存加密结果的切片
	crypted := make([]byte, len(origData))
	// 加密数据
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// EncryptToBase64 加密数据并转换为Base64字符串
// 这是一个便捷方法，将加密后的二进制数据转换为可打印的字符串
// 参数:
//   - origData: 需要加密的原始数据
//   - key: 加密密钥
// 返回:
//   - Base64编码的加密数据和可能的错误
func EncryptToBase64(origData, key []byte) (string, error) {
	// 先进行AES加密
	crypted, err := Encrypt(origData, key)
	if err != nil {
		return "", err
	}
	// 将加密结果转换为Base64字符串（使用URL安全的编码方式）
	return base64.RawURLEncoding.EncodeToString(crypted), nil
}

// Decrypt 使用AES-CBC模式解密数据
// 参数:
//   - crypted: 已加密的数据
//   - key: 解密密钥，必须与加密时使用的密钥相同
// 返回:
//   - 解密后的原始数据和可能的错误
func Decrypt(crypted, key []byte) ([]byte, error) {
	// 创建AES解密块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 获取块大小
	blockSize := block.BlockSize()
	// 创建CBC模式的解密器，IV与加密时相同
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 创建保存解密结果的切片
	origData := make([]byte, len(crypted))
	// 解密数据
	blockMode.CryptBlocks(origData, crypted)
	// 去除填充
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// DecryptFromBase64 从Base64字符串解密数据
// 这是一个便捷方法，用于解密Base64编码的加密数据
// 参数:
//   - data: Base64编码的加密数据
//   - key: 解密密钥
// 返回:
//   - 解密后的原始数据和可能的错误
func DecryptFromBase64(data string, key []byte) ([]byte, error) {
	// 先将Base64字符串解码为字节数组
	crypted, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	// 解密数据
	return Decrypt(crypted, key)
}
