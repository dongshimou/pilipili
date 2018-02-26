package tool

import (
	"fmt"
	"strings"
	"sort"
	"crypto/md5"
	"encoding/hex"
	"io"
	"compress/flate"
	"io/ioutil"
	"compress/gzip"
)

func B_tostring(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
func B_build_url(base string, para map[string]string) string {
	res := base
	res += "?"
	for k, v := range para {
		res += k
		res += "="
		res += v
		res += "&"
	}
	res = strings.TrimRight(res, "&")
	return res
}

//https://interface.bilibili.com/v2/playurl?cid=32253539&appkey=84956560bc028eb7&otype=json&type=&quality=0&qn=0&sign=0d1b3ad4dd857c060d28a31a05d13835
func B_httpBuildQuery(params map[string]string) string {
	list := make([]string, 0, len(params))
	buffer := make([]string, 0, len(params))
	for key := range params {
		list = append(list, key)
	}
	sort.Strings(list)
	for _, key := range list {
		value := params[key]
		buffer = append(buffer, key)
		buffer = append(buffer, "=")
		buffer = append(buffer, value)
		buffer = append(buffer, "&")
	}
	buffer = buffer[:len(buffer)-1]
	return strings.Join(buffer, "")
}
func B_EncodeSign(params map[string]string, secret string) (string, string) {
	queryString := B_httpBuildQuery(params)
	return queryString, B_Md5(queryString + secret)
}

func B_Md5(formal string) string {
	h := md5.New()
	h.Write([]byte(formal))
	return hex.EncodeToString(h.Sum(nil))
}
//deflate解压 Content-Encoding: deflate
func B_flate_decode(in io.ReadCloser) ([]byte, error) {
	reader := flate.NewReader(in)
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

//gzip解压 Content-Encoding: Gzip
func B_Gzip_decode(in io.ReadCloser) ([]byte, error) {
	reader, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func Byte42Uint32(data []byte, big_endian bool) uint32 {
	var i uint32
	if big_endian {
		i = uint32(uint32(data[3]) + uint32(data[2])<<8 + uint32(data[1])<<16 + uint32(data[0])<<24)
	}else{
		i = uint32(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24)
	}
	return i
}

func Byte32Uint32(data []byte, big_endian bool) uint32 {
	var i uint32
	if big_endian {
		i = uint32(uint32(data[2]) + uint32(data[1])<<8 + uint32(data[0])<<16)
	}else{
		i = uint32(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16)
	}
	return i
}