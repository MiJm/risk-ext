package views

import (
	"risk-ext/models"
	"strconv"
	"time"

	"github.com/kataras/iris"
)

type SimView struct {
	Views
}

func (this *SimView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		//		"GET": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_SERVICE, MANAGER_ASSISTANT}}
	}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *SimView) Get(ctx iris.Context) (statusCode int, data interface{}) {
	err, bill := this.GetBillGroup()
	if err == nil {
		for _, v := range bill.Data {
			code := v.Bg_code
			page := 1
			size := 1000
			err, list := this.SimList(code, page, size)
			time.Sleep(time.Second * 4)
			if err == nil {
				if list.Code == 200 {
					total := list.Num_pages
					for ; page <= total; page++ {
						time.Sleep(time.Second * 4)
						err, list1 := this.SimList(code, page, size)
						if err != nil {
							continue
						}
						datas := list1.Data
						for _, k := range datas {
							simCard, err := strconv.Atoi(k.Msisdn)
							if err != nil {
								continue
							}
							dev, err := new(models.Devices).OneBySim(uint64(simCard))
							if err != nil {
								continue
							}
							dev.Device_sim_info = k
							err = dev.Update(false)
							if err != nil {
								continue
							}
						}
					}
				}
			}
		}
	}
	return
}
