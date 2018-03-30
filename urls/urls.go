package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	app.AddPath("v2/amount/", new(views.AmountView))
}
