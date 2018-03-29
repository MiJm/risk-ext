package urls

type M map[string]interface{}

import (
	"risk-ext/app"
	"risk-ext/models"
	"risk-ext/views"
)

const (
    MANAGER_ADMIN     = 0
	MANAGER_SERVICE   = 1
	MANAGER_STORE     = 2
	MANAGER_ASSISTANT = 3

	MEMBER_SUPER      = 2
	MEMBER_ADMIN      = 1
	MEMBER_GENERAL    = 0
)

var s = int[2][]{{2},{1}}
func init() {
	app.AddPath("user", new(views.UsersView))
}
