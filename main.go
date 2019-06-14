package main

import (
	"risk-ext/app"
	_ "risk-ext/urls"
)

func main() {
	app.StartUdp()
	app.Run()
}
