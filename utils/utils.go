package utils

import (
	"os"
	"reflect"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

//字符串时间转为时间戳
func Str2Time(datetime string) uint32 {
	timeLayout := "2006-01-02 15:04:05"
	timeLayout1 := "2006/01/02 15:04:05"
	timeLayout2 := "2006-1-2 15:4:5"
	timeLayout3 := "2006/1/2 15:4:5"
	loc, _ := time.LoadLocation("Local")
	datetime = strings.TrimSpace(datetime)
	times, err := time.ParseInLocation(timeLayout, datetime, loc)
	if err != nil {
		times, err = time.ParseInLocation(timeLayout1, datetime, loc)
		if err != nil {
			times, err = time.ParseInLocation(timeLayout2, datetime, loc)
			if err != nil {
				times, err = time.ParseInLocation(timeLayout3, datetime, loc)
				if err != nil {
					return uint32(0)
				}
				return uint32(times.Unix())
			}
			return uint32(times.Unix())
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

//结构体转为map
//obj要转换的strut
//noCom 是否不匹配空值
func Struct2Map(obj interface{}, noCom ...bool) bson.M {

	t := reflect.TypeOf(obj)
	emptyObj := reflect.New(t).Elem().Interface() //获取空类型结构体
	v := reflect.ValueOf(obj)
	nv := reflect.ValueOf(emptyObj)
	var data = bson.M{}
	for i := 0; i < t.NumField(); i++ {
		key := strings.ToLower(t.Field(i).Name)
		if len(noCom) > 0 && noCom[0] {
			data[key] = v.Field(i).Interface()
		} else if !reflect.DeepEqual(nv.Field(i).Interface(), v.Field(i).Interface()) {
			data[key] = v.Field(i).Interface()
		}
	}
	return data

}
