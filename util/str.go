package util

import (
	"sort"
)

// StrInSlice 判断指定字符是否在切片中
// https://blog.51cto.com/u_15023263/2558348
func StrInSlice(target string, strSlice []string) bool {
	if strSlice == nil {
		return false
	}
	sort.Strings(strSlice)
	index := sort.SearchStrings(strSlice, target)
	if index < len(strSlice) && strSlice[index] == target {
		return true
	}
	return false
}
