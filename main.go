package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"tieba-reconstructor/config"
	"tieba-reconstructor/controller"

	"github.com/beego/beego/v2/server/web"
)

const (
	configDir string = "config.json"
)

var Config *config.Config

func loadConfig() {
	file, err := ioutil.ReadFile(configDir)
	Config = &config.Config{}
	if err != nil {
		os.Create(configDir)
	} else {
		json.Unmarshal(file, Config)
	}
}
func main() {
	controller.Init()
	web.SetStaticPath("/static", "static")

	web.Run()
}
