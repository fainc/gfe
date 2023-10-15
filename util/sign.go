package util

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/util/gconv"
)

type SignatureStrKeyOptions struct {
	Key   string
	Value string
}

// SignatureStr 常用签名处理方法 map转签名串（字典升序），适用一层的map数据
func SignatureStr(m map[string]string, keyOptions *SignatureStrKeyOptions, exclude []string) (str string) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	// 按字典升序排列
	sort.Strings(keys)
	str = ""
	for _, k := range keys {
		if _, ok := m[k]; ok {
			if exclude == nil || !StrInSlice(k, exclude) {
				str = fmt.Sprintf("%s&%s=%s", str, k, m[k])
			}
		}
	}
	if keyOptions != nil && keyOptions.Key != "" && keyOptions.Value != "" {
		str = fmt.Sprintf("%s&%s=%s", str, keyOptions.Key, keyOptions.Value)
	}
	str = strings.TrimPrefix(str, "&")
	return
}

// SignatureComplexStr 常用签名处理方法 map转签名串（字典升序），适用多层的复杂map数据
func SignatureComplexStr(m map[string]interface{}, keyOptions *SignatureStrKeyOptions, exclude []string) (str string) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	// 按字典升序排列
	sort.Strings(keys)
	str = ""
	for _, k := range keys {
		if _, ok := m[k]; ok {
			v := gconv.String(m[k])
			is := json.Valid([]byte(v)) // 判断是否JSON
			if is {
				var vj map[string]interface{}
				err := json.Unmarshal([]byte(v), &vj) // 尝试以对象解析
				if err == nil {
					v = SignatureComplexStr(vj, nil, nil)
				} else {
					var vjao []map[string]interface{}
					err = json.Unmarshal([]byte(v), &vjao) // 尝试以数组对象解析
					if err == nil {
						v = arrayObjectHandler(vjao, nil, nil)
					} else {
						var vja []interface{}
						err = gjson.Unmarshal([]byte(v), &vja) // 尝以纯数组解析
						if err == nil {
							v = arrayHandler(vja, nil)
						}
					}
				}
			}
			if exclude == nil || !StrInSlice(k, exclude) {
				str = fmt.Sprintf("%s&%s=%s", str, k, v)
			}
		}
	}
	if keyOptions != nil && keyOptions.Key != "" && keyOptions.Value != "" {
		str = fmt.Sprintf("%s&%s=%s", str, keyOptions.Key, keyOptions.Value)
	}
	str = strings.TrimPrefix(str, "&")
	return
}

// arrayObjectHandler 数组对象签名方法
func arrayObjectHandler(m []map[string]interface{}, keyOptions *SignatureStrKeyOptions, exclude []string) (str string) {
	str = ""
	for mk, mv := range m {
		s := SignatureComplexStr(mv, keyOptions, exclude)
		str = fmt.Sprintf("%s&%v=%s", str, mk, s)
	}
	if keyOptions != nil && keyOptions.Key != "" && keyOptions.Value != "" {
		str = fmt.Sprintf("%s&%s=%s", str, keyOptions.Key, keyOptions.Value)
	}
	str = strings.TrimPrefix(str, "&")
	return
}

// arrayHandler 数组签名方法
func arrayHandler(m []interface{}, keyOptions *SignatureStrKeyOptions) (str string) {
	str = ""
	for mk, mv := range m {
		ms := gconv.String(mv)
		is := gjson.Valid(ms) // 判断是否JSON
		if is {
			var v map[string]interface{}
			err := json.Unmarshal([]byte(ms), &v)
			if err == nil {
				ms = SignatureComplexStr(v, nil, nil)
			} else {
				var v2 []interface{}
				err = json.Unmarshal([]byte(ms), &v2)
				if err == nil {
					ms = arrayHandler(v2, keyOptions)
				}
			}
		}
		str = fmt.Sprintf("%s&%v=%s", str, mk, ms)
	}
	if keyOptions != nil && keyOptions.Key != "" && keyOptions.Value != "" {
		str = fmt.Sprintf("%s&%s=%s", str, keyOptions.Key, keyOptions.Value)
	}
	str = strings.TrimPrefix(str, "&")
	return
}
