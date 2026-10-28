// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tool

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"strconv"
	"unicode"

	"github.com/unknwon/com"
)

// MD5Bytes encodes string to MD5 bytes.
func MD5Bytes(str string) []byte {
	m := md5.New()
	m.Write([]byte(str))
	return m.Sum(nil)
}

// MD5 encodes string to MD5 hex value.
func MD5(str string) string {
	return hex.EncodeToString(MD5Bytes(str))
}

// SHA1 encodes string to SHA1 hex value.
func SHA1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// ShortSHA1 truncates SHA1 string length to at most 10.
func ShortSHA1(sha1 string) string {
	if len(sha1) > 10 {
		return sha1[:10]
	}
	return sha1
}

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomString returns generated random string in given length of characters.
// It also returns possible error during generation.
func RandomString(n int) (string, error) {
	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))

	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer), nil
}

func randomInt(max *big.Int) (int, error) {
	rand, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return int(rand.Int64()), nil
}

func RandomMailCode(n int) (string, error) {
	result := ""
	max := big.NewInt(int64(10))

	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}
		result += strconv.Itoa(index)
	}

	return result, nil
}

// Subtract deals with subtraction of all types of number.
func Subtract(left interface{}, right interface{}) interface{} {
	var rleft, rright int64
	var fleft, fright float64
	var isInt = true
	switch left.(type) {
	case int:
		rleft = int64(left.(int))
	case int8:
		rleft = int64(left.(int8))
	case int16:
		rleft = int64(left.(int16))
	case int32:
		rleft = int64(left.(int32))
	case int64:
		rleft = left.(int64)
	case float32:
		fleft = float64(left.(float32))
		isInt = false
	case float64:
		fleft = left.(float64)
		isInt = false
	}

	switch right.(type) {
	case int:
		rright = int64(right.(int))
	case int8:
		rright = int64(right.(int8))
	case int16:
		rright = int64(right.(int16))
	case int32:
		rright = int64(right.(int32))
	case int64:
		rright = right.(int64)
	case float32:
		fright = float64(left.(float32))
		isInt = false
	case float64:
		fleft = left.(float64)
		isInt = false
	}

	if isInt {
		return rleft - rright
	} else {
		return fleft + float64(rleft) - (fright + float64(rright))
	}
}

// EllipsisString returns a truncated short string,
// it appends '...' in the end of the length of string is too large.
func EllipsisString(str string, length int) string {
	if len(str) < length {
		return str
	}
	return str[:length-3] + "..."
}

// TruncateString returns a truncated string with given limit,
// it returns input string if length is not reached limit.
func TruncateString(str string, limit int) string {
	if len(str) < limit {
		return str
	}
	return str[:limit]
}

// StringsToInt64s converts a slice of string to a slice of int64.
func StringsToInt64s(strs []string) []int64 {
	ints := make([]int64, len(strs))
	for i := range strs {
		ints[i] = com.StrTo(strs[i]).MustInt64()
	}
	return ints
}

// Int64sToStrings converts a slice of int64 to a slice of string.
func Int64sToStrings(ints []int64) []string {
	strs := make([]string, len(ints))
	for i := range ints {
		strs[i] = com.ToStr(ints[i])
	}
	return strs
}

// Int64sToMap converts a slice of int64 to a int64 map.
func Int64sToMap(ints []int64) map[int64]bool {
	m := make(map[int64]bool)
	for _, i := range ints {
		m[i] = true
	}
	return m
}

// IsLetter reports whether the rune is a letter (category L).
// https://github.com/golang/go/blob/master/src/go/scanner/scanner.go#L257
func IsLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func StrSliceContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func StrSliceRemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func SHA256Encode(str string) (string, error) {
	h := sha256.New()
	h.Write([]byte(str))
	res := hex.EncodeToString(h.Sum(nil))
	return res, nil
}

func StringToPointer(str string) *string {
	return &str
}

func BoolToPointer(flag bool) *bool {
	return &flag
}

func NoInjection(i interface{}) bool {
	return i == nil
}

func SliceCompare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !StrSliceContains(b, a[i]) {
			return false
		}
	}
	return true
}
