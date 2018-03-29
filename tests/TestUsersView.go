package tests

import (
	"risk-ext/app"
	"testing"

	"github.com/kataras/iris/httptest"
)

func TestUserView(t *testing.T) {
	app := app.App()
	e := httptest.New(t, app)
	e.GET("/user").Expect().Status(httptest.StatusOK)
}
