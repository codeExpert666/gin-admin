// Package rand 提供了安全的随机字符串生成功能
package rand

import (
	"bytes"
	"crypto/rand"
	"errors"
)

// 定义随机字符串生成的标志位，使用位运算来组合不同的字符类型
const (
	Ldigit             = 1 << iota                        // 仅数字
	LlowerCase                                            // 仅小写字母
	LupperCase                                            // 仅大写字母
	LlowerAndUpperCase = LlowerCase | LupperCase          // 小写和大写字母
	LdigitAndLowerCase = Ldigit | LlowerCase              // 数字和小写字母
	LdigitAndUpperCase = Ldigit | LupperCase              // 数字和大写字母
	LdigitAndLetter    = Ldigit | LlowerCase | LupperCase // 数字和所有字母
)

// 定义可用于生成随机字符串的字符集
var (
	digits           = []byte("0123456789")                 // 数字字符集
	lowerCaseLetters = []byte("abcdefghijklmnopqrstuvwxyz") // 小写字母字符集
	upperCaseLetters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ") // 大写字母字符集
)

// 定义错误类型
var (
	ErrInvalidFlag = errors.New("invalid flag") // 当传入的标志位无效时返回此错误
)

// Random 生成指定长度和类型的随机字符串
// length: 要生成的随机字符串的长度
// flag: 指定随机字符串包含的字符类型（数字、小写字母、大写字母的组合）
// 返回生成的随机字符串和可能的错误
func Random(length, flag int) (string, error) {
	// 如果长度小于1，设置默认长度为6
	if length < 1 {
		length = 6
	}

	// 根据标志位获取源字符集
	source, err := getFlagSource(flag)
	if err != nil {
		return "", err
	}

	// 生成随机字节序列
	b, err := randomBytesMod(length, byte(len(source)))
	if err != nil {
		return "", err
	}

	// 将随机字节转换为字符串
	var buf bytes.Buffer
	for _, c := range b {
		buf.WriteByte(source[c])
	}

	return buf.String(), nil
}

// getFlagSource 根据标志位组合相应的字符集
// flag: 标志位
// 返回组合后的字符集和可能的错误
func getFlagSource(flag int) ([]byte, error) {
	var source []byte

	// 根据标志位使用位运算来判断需要包含哪些字符集
	if flag&Ldigit > 0 {
		source = append(source, digits...)
	}

	if flag&LlowerCase > 0 {
		source = append(source, lowerCaseLetters...)
	}

	if flag&LupperCase > 0 {
		source = append(source, upperCaseLetters...)
	}

	// 如果没有选择任何字符集，返回错误
	sourceLen := len(source)
	if sourceLen == 0 {
		return nil, ErrInvalidFlag
	}
	return source, nil
}

// randomBytesMod 生成指定长度的随机字节序列，并对每个字节进行取模运算
// length: 需要生成的字节序列长度
// mod: 用于取模的值，通常是源字符集的长度
// 返回生成的字节序列和可能的错误
func randomBytesMod(length int, mod byte) ([]byte, error) {
	b := make([]byte, length)
	// 计算不会产生模偏差的最大值
	max := 255 - 255%mod
	i := 0

LROOT:
	for {
		// 生成随机字节，多生成一些以应对需要跳过的数字
		r, err := randomBytes(length + length/4)
		if err != nil {
			return nil, err
		}

		// 对每个随机字节进行处理
		for _, c := range r {
			if c >= max {
				// 跳过可能导致模偏差的数字
				continue
			}

			// 取模运算得到最终的随机索引
			b[i] = c % mod
			i++
			if i == length {
				break LROOT
			}
		}
	}

	return b, nil
}

// randomBytes 使用密码学安全的随机数生成器生成随机字节序列
// length: 需要生成的字节序列长度
// 返回生成的随机字节序列和可能的错误
func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	// 使用crypto/rand包生成随机字节
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
