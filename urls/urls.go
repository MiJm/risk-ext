package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	app.AddPath("v2/amount/", new(views.AmountView))
	app.AddPath("v2/track/", new(views.TrackView))
	app.AddPath("v2/log/", new(views.LogView))
}
