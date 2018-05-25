package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"image"
	"image/draw"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/disintegration/imaging"

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

//缩放图片
func ImageResize(src, savePath string) (err error) {
	im, err := imaging.Open(src)
	if err != nil {
		errors.New("文件打开失败")
	}
	width := im.Bounds().Size().X
	height := im.Bounds().Size().Y

	//	if width <= 300 && height <= 300 { //小图不裁剪
	//		return errors.New("hh")
	//	}

	var m *image.NRGBA
	var dstWidth int
	var desheigh int

	if width > height {
		if height <= 300 {
			dstWidth = height
			desheigh = height

			m = imaging.Resize(im, 0, height, imaging.CatmullRom)
		} else {
			dstWidth = 300
			desheigh = 300
			width = width * 300 / height
			height = 300
			m = imaging.Resize(im, 0, 300, imaging.CatmullRom)
		}

	} else {
		if width <= 300 {
			dstWidth = width
			desheigh = width
			m = imaging.Resize(im, width, 0, imaging.CatmullRom)
		} else {
			dstWidth = 300
			desheigh = 300
			height = height * 300 / width
			width = 300
			m = imaging.Resize(im, 300, 0, imaging.CatmullRom)
		}

	}

	jpg := image.NewRGBA(image.Rect(0, 0, dstWidth, desheigh))
	pio := new(image.Point)
	pio.X = (width - dstWidth) / 2
	pio.Y = (height - desheigh) / 2
	draw.Draw(jpg, m.Bounds().Add(image.Pt(0, 0)), m, *pio, draw.Src)
	out, err := os.Create(savePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	//	imaging.Encode(out, m, imaging.PNG)
	imaging.Encode(out, jpg, imaging.PNG)
	return
}

//字符串转md5
func String2Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str)) // 需要加密的字符串为 123456
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr) // 输出加密结果
}

func SubString(str string, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)
	begin := lth - length
	if length >= lth {
		begin = 0
	}
	// 返回子串
	return string(rs[begin:lth])
}
