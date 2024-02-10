//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"time"
)

func main_ui() {
	LogInfo("PalServer 守护进程 v" + AppVserion)
	// 启动Pal服务器
	if !startPalServer() {
		LogInfo("服务器启动失败")
		LogInfo("Pal 服务器启动失败, 请检查配置文件和服务器启动命令是否正确")
		os.Exit(1)
	}
	GoGuard()

	commandLoop()
}

func commandLoop() {
	for {
		var input string
		fmt.Print("输入指令: ")
		fmt.Scanln(&input)
		if input == "stop" {
			LogInfo("正在保存游戏存档并退出程序...")
			ExecRconCmd("Save")
			time.Sleep(5 * time.Second)
			ExecRconCmd("DoExit")
			LogInfo("已经退出程序")
			os.Exit(0)
		}
		if input == "reloadcfg" {
			LogInfo("重新加载配置文件...")
			InitConfig()
			LogInfo("配置文件已重新加载")
			continue
		}
		if input == "backup" {
			backupPalServer()
			continue
		}
		if input == "restart" {
			restartPalServer(false)
			continue
		}
		resp := ExecRconCmd(input)
		if resp == "" {
			LogError("发送指令失败")
		} else {
			LogInfo("[Rcon]: ", resp)
		}
	}
}
func flush2Text(data string) {}
