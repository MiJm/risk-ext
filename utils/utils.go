package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"

	socketgo "github.com/nulijiabei/socketgo"
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
		//key := strings.ToLower(t.Field(i).Name)
		key := t.Field(i).Tag.Get("json")
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

//将时间秒转天数
func Timelen(intTimer int) (timeStr string) {
	var day, hour, minute int
	if intTimer > 24*3600 {
		day = intTimer / (24 * 3600)
		Dremain := intTimer % (24 * 3600)
		if Dremain > 3600 {
			hour = Dremain / 3600
			Hremain := Dremain % 3600
			if Hremain > 60 {
				minute = Hremain / 60
			}
		} else {
			minute = Dremain / 60
		}
	} else {
		if intTimer > 3600 {
			hour = intTimer / 3600
			Hremain := intTimer % 3600
			if Hremain > 60 {
				minute = Hremain / 60
			}
		} else {
			minute = intTimer / 60
		}
	}

	if day > 0 {
		timeStr = timeStr + strconv.Itoa(day) + "天"
	}
	if hour > 0 {
		timeStr = timeStr + strconv.Itoa(hour) + "小时"
	}
	if minute > 0 {
		timeStr = timeStr + strconv.Itoa(minute) + "分钟"
	}
	if day == 0 && hour == 0 && minute == 0 {
		timeStr = "小于1分钟"
	}
	return timeStr
}

func Time2Str1(datetime uint32) string {
	timeLayout := "2006-01-02"
	tm := time.Unix(int64(datetime), 0)
	return tm.Format(timeLayout)
}

/***
 * 加密算法开始
 */
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func Aes128Encrypt(origData, key []byte, IV []byte) ([]byte, error) {
	if key == nil || len(key) != 16 {
		return nil, nil
	}
	if IV != nil && len(IV) != 16 {
		return nil, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, IV[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func Aes128Decrypt(crypted, key []byte, IV []byte) ([]byte, error) {
	if key == nil || len(key) != 16 {
		return nil, nil
	}
	if IV != nil && len(IV) != 16 {
		return nil, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, IV[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

/**
 * AES 128解密方法（目前用于微信敏感信息解密）
 * src 需要解密的字符串
 * sKey session key（会话key）
 * iv 二级钥匙
 */
func PswDecrypt(src, sKey, iv string) (string, error) {
	var result []byte
	var err error
	asKey, err := base64.StdEncoding.DecodeString(sKey)
	if err != nil {
		return "", err
	}
	aiv, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}
	result, err = base64.RawStdEncoding.DecodeString(src)
	if err != nil {
		result, err = base64.StdEncoding.DecodeString(src)
		if err != nil {
			return "", err
		}
	}
	origData, err := Aes128Decrypt(result, asKey, aiv)
	if err != nil {
		return "", err
	}
	return string(origData), nil
}

//解密入口
func AesDecode(str string) (rs string, err error) {
	var aeskey = []byte("joygin1234567890")
	bytesPass, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(err)
		return
	}
	tpass, err := AesDecrypt(bytesPass, aeskey)
	if err != nil {
		fmt.Println(err)
		return
	}
	rs = string(tpass)
	return
}

/***
 * 加密算法结束
 */

//发送实时指令
//devId设备ID
//Type指令类型  0=重启设备 1=断油电 2=通油电 3开追踪 4设置闹钟 5设置定时 6设置星期 7设防 8撤防 9寻车
//args 指令参数 不填为空
//
func ExeCmd(devId uint64, Type uint8, args string) bool {
	var rs = make(chan bool, 1)
	var cmd = struct {
		Imei  uint64 `json:"cmd_imei"`
		Type  uint8  `json:"cmd_type"` //指令类型  0=重启设备 1=断油电 2=通油电 3开追踪 4设置闹钟 5设置定时 6设置星期 7设防 8撤防 9寻车
		Args  string `json:"cmd_args"`
		CmdId string `json:"cmd_id"`
	}{devId, Type, args, ""}

	err := socketgo.NewTCP("cmdServer", "1985", 3).ReadWrite(func(conn *net.TCPConn) error {
		mjson, err := json.Marshal(cmd)
		if err != nil {
			rs <- false
			return nil
		}

		_, err = conn.Write([]byte("execmd|" + string(mjson) + "$"))

		if err != nil {
			rs <- false
			return nil
		}
		var buf = make([]byte, 128)

		blen, err := conn.Read(buf)
		if err != nil {
			rs <- false
			return nil
		}
		buf = buf[:blen]
		if string(buf) != "ok" {
			rs <- false
			return nil
		}

		rs <- true
		return nil
	})

	if err != nil {
		return false
	} else {
		return <-rs
	}
}
