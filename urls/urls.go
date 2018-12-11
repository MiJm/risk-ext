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

	app.AddPath("v2/amount/", amountView)
	app.AddPath("v2/log/", logView)
	app.AddPath("v2/reports/", reportView)
	app.AddPath("v2/reports/{report_id}", reportView)
	app.AddPath("v2/shares/", shareView)
	app.AddPath("v2/shares/{share_id}", shareView)
}
