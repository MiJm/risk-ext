package main

import (
	"risk-ext/app"
	_ "risk-ext/urls"
)

func main() {
	app.StartUdp("1992")
	app.Run()
}
