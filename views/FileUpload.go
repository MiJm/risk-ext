package views

import (
	"fmt"
	"io"
	"os"
	"risk-ext/app"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
	"time"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type FileUploadView struct {
	Views
}

func (this *FileUploadView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"CUSTOMER": A{1}},
		"GET":    MA{"CUSTOMER": A{1}},
		"POST":   MA{"CUSTOMER": A{1}},
		"DELETE": MA{"CUSTOMER": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *FileUploadView) Post(ctx iris.Context) (statuCode int, data app.M) {
	statuCode = 400
	data = make(app.M)
	img, _, err := ctx.FormFile("img")
	if err != nil {
		data["code"] = 0
		data["msg"] = "上传图片失败"
		data["data"] = nil
		return
	}
	defer img.Close()
	Title := fmt.Sprintf("%s.png", bson.NewObjectId().Hex())
	SaveUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/"
	OpenUrl := "photos/" + time.Now().Format("200601") + "/"
	err = utils.IsFile(SaveUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	f, err := os.OpenFile(SaveUrl+Title, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["msg"] = err.Error()
		data["data"] = nil
		return
	}
	defer f.Close()
	_, err = io.Copy(f, img)
	if err != nil {
		data["code"] = 0
		data["msg"] = "保存文件失败"
		data["data"] = nil
		return
	}
	ThumbUrl := config.GetString("WarrantyFiles") + time.Now().Format("200601") + "/thumb"
	ThumbOpenUrl := "photos/" + time.Now().Format("200601") + "/thumb"
	err = utils.IsFile(ThumbUrl)
	if err != nil {
		data["code"] = 0
		data["msg"] = "创建文件失败"
		data["data"] = nil
		return
	}
	err = utils.ImageResize(SaveUrl+Title, ThumbUrl+Title)
	if err != nil {
		data["code"] = 0
		data["msg"] = "生成文件失败"
		data["data"] = nil
		return
	}
	statuCode = 200
	data["code"] = 1
	data["msg"] = "OK"
	var pathObj models.ImgPath
	pathObj.Path = OpenUrl + Title
	pathObj.ThumbPath = ThumbOpenUrl + Title
	data["data"] = pathObj
	return

}

//获取详情或列表待用
func (this *FileUploadView) Get(ctx iris.Context) (statuCode int, data app.M) {
	return
}

//更新操作待用
func (this *FileUploadView) Put(ctx iris.Context) (statuCode int, data app.M) {
	return
}

//删除操作待用
func (this *FileUploadView) Delete(ctx iris.Context) (statuCode int, data app.M) {
	return
}
