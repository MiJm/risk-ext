package views

import (
	"fmt"
	"risk-ext/models"

	"github.com/kataras/iris"
)

const (
	MANAGER_ADMIN     = 0
	MANAGER_SERVICE   = 1
	MANAGER_STORE     = 2
	MANAGER_ASSISTANT = 3

	MEMBER_SUPER   = 2
	MEMBER_ADMIN   = 1
	MEMBER_GENERAL = 0
)

type A []uint8
type MA map[string]A
type PMS map[string]MA

var Session *models.Session

type Views struct {
}

func (this *Views) Auth(ctx iris.Context) bool {
	if Session == nil {
		token := ctx.GetHeader("token")
		Session = new(models.Session).Data(token)
	}
	return true
}

func (this *Views) CheckPerms(perm MA) bool {
	if perm == nil {
		return false //方法为授权默认无权限访问
	}

	if perm["NOLOGIN"] != nil {
		return true //无需登录
	}

	if Session == nil {
		return false //session不存在无权限访问
	}
	fmt.Println(perm)
	if Session.Type == 1 { //当前登录用户为普通用户
		if perm["USER"] == nil {
			return false //普通用户无权限
		} else {
			for _, p := range perm["USER"] {
				if p == Session.User.UserLevel {
					return true //找到了相应的权限
				}
			}
			return false //未找到相应权限
		}
	} else { //后台管理员
		if perm["ADMIN"] == nil {
			return false //管理员无权限
		} else {
			for _, p := range perm["ADMIN"] {
				if p == Session.Manager.Manager_level {
					return true //找到了相应的权限
				}
			}
			return false //未找到相应权限
		}
	}

}
