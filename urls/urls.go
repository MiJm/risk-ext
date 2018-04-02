package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	app.AddPath("v2/amount/", new(views.AmountView))
	app.AddPath("v2/log/", new(views.LogView))
	app.AddPath("v2/reports/", new(views.ReportView))
	app.AddPath("v2/reports/report_id", new(views.ReportView))
}
