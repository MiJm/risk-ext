package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	amountView := new(views.AmountView)
	logView := new(views.LogView)
	reportView := new(views.ReportView)
	shareView := new(views.SharesView)
	commonView := new(views.CommonsView)
	simView := new(views.SimView)

	app.AddPath("v2/amount/", amountView)
	app.AddPath("v2/log/", logView)
	app.AddPath("v2/reports/", reportView)
	app.AddPath("v2/reports/{report_id}", reportView)
	app.AddPath("v2/shares/", shareView)
	app.AddPath("v2/shares/{params}", shareView)
	app.AddPath("v2/commons/", commonView)
	app.AddPath("v2/commons/{report_id}", commonView)
	app.AddPath("v2/sim/", simView)
}
