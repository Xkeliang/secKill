package main

import (
	"encoding/json"
	"fmt"
	"github.com/beego/logs"
)

func converLogLever(lever string) int {
	switch lever {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}

	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = appConfig.LogPath
	config["lever"] = converLogLever(appConfig.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshar failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	return
}
