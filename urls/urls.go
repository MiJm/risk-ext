package urls

import (
	"risk-ext/app"
	"risk-ext/views"
)

func init() {
	app.AddPath("user", new(views.UsersView))
}
