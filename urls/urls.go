package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	amountView := new(views.AmountView)
	logView := new(views.LogView)
	reportView := new(views.ReportView)

	app.AddPath("v2/amount/", amountView)
	app.AddPath("v2/log/", logView)
	app.AddPath("v2/reports/", reportView)
	app.AddPath("v2/reports/{report_id}", reportView)
}
