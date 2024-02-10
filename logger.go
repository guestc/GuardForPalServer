package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

type LogColor struct {
	Info *color.Color
	Warn *color.Color
	Err  *color.Color
	Dbg  *color.Color
}

var logColor = LogColor{
	Info: color.New(color.Reset),
	Warn: color.New(color.FgHiWhite).Add(color.BgYellow),
	Err:  color.New(color.FgHiWhite).Add(color.BgRed),
	Dbg:  color.New(color.FgHiCyan),
}

func log(data string, lc *color.Color) {
	msg := fmt.Sprintf("[%s] %s", getDate(), data)
	flush2Text(msg)
	//fmt.Println(msg)
	lc.Fprintf(os.Stdout, msg+"\n")
}

// LogInfo 打印信息
func LogInfo(datas ...any) {
	sb := ""
	for _, data := range datas {
		if sb != "" {
			sb += " "
		}
		sb += fmt.Sprintf("%v", data)
	}
	log(fmt.Sprintf("[Info]: %s", sb), logColor.Info)
}

// LogWarning 打印警告信息
func LogWarning(datas ...any) {
	sb := ""
	for _, data := range datas {
		if sb != "" {
			sb += " "
		}
		sb += fmt.Sprintf("%v", data)
	}
	log(fmt.Sprintf("[Warning]: %s", sb), logColor.Warn)
}

// LogError 打印错误信息
func LogError(datas ...any) {
	sb := ""
	for _, data := range datas {
		if sb != "" {
			sb += " "
		}
		sb += fmt.Sprintf("%v", data)
	}
	log(fmt.Sprintf("[Error]: %v", sb), logColor.Err)
}

// LogDebug 打印调试信息
func LogDebug(datas ...any) {
	if !Cfg.Server.Debug {
		return
	}
	sb := ""
	for _, data := range datas {
		if sb != "" {
			sb += " "
		}
		sb += fmt.Sprintf("%v", data)
	}
	log(fmt.Sprintf("[Debug]: %v", sb), logColor.Dbg)
}

// 当前时间 格式化为 YYYY-MM-DD HH:MM:SS
func getDate() string {
	return time.Now().Format("01-02 15:04:05")
}
