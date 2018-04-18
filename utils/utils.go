package utils

import (
	"os"
	"strings"
	"time"
)

//字符串时间转为时间戳
func Str2Time(datetime string) uint32 {
	timeLayout := "2006-01-02 15:04:05"
	timeLayout1 := "2006/01/02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	datetime = strings.TrimSpace(datetime)
	times, err := time.ParseInLocation(timeLayout, datetime, loc)
	if err != nil {
		times, err = time.ParseInLocation(timeLayout1, datetime, loc)
		if err != nil {
			return uint32(0)
		}
		return uint32(times.Unix())
	}
	return uint32(times.Unix())
}

//判断文件夹是否存在
func IsFile(name string) (err error) {
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		err = os.MkdirAll(name, os.ModePerm)
	}
	return
}
