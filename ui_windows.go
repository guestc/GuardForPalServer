//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
)

var LogEntry *widget.Entry
var Bd_LogEntry binding.String
var MyApp fyne.App

// 定义一个读写互斥锁
var rwMutex sync.RWMutex

func flush2Text(data string) {
	if !IsInWindows {
		return
	}
	if LogEntry == nil {
		return
	}
	// 读写锁

	go func() {
		rwMutex.RLock()
		content := LogEntry.Text
		if len(content) > 1024*1024 {
			content = "-----------已清除日志-----------"
		}
		LogEntry.SetText(fmt.Sprintf("%s\n%s", content, data))
		LogEntry.CursorRow = len(LogEntry.Text) - 1
		time.Sleep(time.Millisecond * 100)
		LogEntry.Refresh()
		rwMutex.RUnlock()
	}()
}
func setZh() {
	fonts := findfont.List()
	for _, font := range fonts {
		if strings.Contains(font, "simsun.ttc") || strings.Contains(font, "simhei.ttf") || strings.Contains(font, "msyh.ttf") || strings.Contains(font, "simkai.ttf") {
			LogInfo("Found font: ", font)
			os.Setenv("FYNE_FONT", font)
			os.Setenv("FYNE_FONT_MONOSPACE", font) // 注意：如果不设置这个环境变量会导致 widget.TextGrid 中的中文乱码
			break
		}
	}
}

func main_ui() {
	setZh()
	MyApp = app.New()
	myWindow := MyApp.NewWindow("PalServer 守护进程 v" + AppVserion)
	myWindow.Resize(fyne.NewSize(850, 550))

	// Server Tab
	bd_execstart := binding.NewString()
	bd_execstart.Set(Cfg.Server.ExecStart)
	bd_rcon_ip := binding.NewString()
	bd_rcon_ip.Set(Cfg.Server.RconIP)
	bd_rcon_port := binding.NewInt()
	bd_rcon_port.Set(Cfg.Server.RconPort)
	bd_rcon_pass := binding.NewString()
	bd_rcon_pass.Set(Cfg.Server.RconPass)
	bd_debug := binding.NewBool()
	bd_debug.Set(Cfg.Server.Debug)

	serverTab := container.NewVBox(
		widget.NewLabel("服务器启动路径"),
		widget.NewEntryWithData(bd_execstart),
		widget.NewLabel("Rcon IP地址"),
		widget.NewEntryWithData(bd_rcon_ip),
		widget.NewLabel("Rcon 端口"),
		widget.NewEntryWithData(binding.IntToString(bd_rcon_port)),
		widget.NewLabel("Rcon 密码"),
		widget.NewEntryWithData(bd_rcon_pass),
		widget.NewCheckWithData("调试模式", bd_debug),

		widget.NewButton("保存", func() {
			// 保存Server配置的逻辑
			var err error
			Cfg.Server.ExecStart, err = bd_execstart.Get()
			if err != nil {
				LogError("bd_execstart.Get() failed: ", err)
			}
			Cfg.Server.RconIP, err = bd_rcon_ip.Get()
			if err != nil {
				LogError("bd_rcon_ip.Get() failed: ", err)
			}
			Cfg.Server.RconPort, err = bd_rcon_port.Get()
			if err != nil {
				LogError("bd_rcon_port.Get() failed: ", err)
			}
			Cfg.Server.RconPass, err = bd_rcon_pass.Get()
			if err != nil {
				LogError("bd_rcon_pass.Get() failed: ", err)
			}
			Cfg.Server.Debug, err = bd_debug.Get()
			if err != nil {
				LogError("bd_debug.Get() failed: ", err)
			}
			LogDebug("ServerConfig: ", Cfg.Server)
			SaveConfig()
		}),
	)

	// Backup Tab
	bd_backup_dir := binding.NewString()
	bd_backup_dir.Set(Cfg.Backup.BackupDir)
	bd_saved_dir := binding.NewString()
	bd_saved_dir.Set(Cfg.Backup.SavedDir)
	bd_backup_interval := binding.NewString()
	bd_backup_interval.Set(Cfg.Backup.BackupInterval)
	bd_backup_max_count := binding.NewInt()
	bd_backup_max_count.Set(Cfg.Backup.BackupMaxCount)
	bd_backup_max_overwrite := binding.NewBool()
	bd_backup_max_overwrite.Set(Cfg.Backup.BackupMaxOverwrite)
	bd_backup_compress := binding.NewBool()
	bd_backup_compress.Set(Cfg.Backup.BackupCompress)
	bd_backup_enable := binding.NewBool()
	bd_backup_enable.Set(Cfg.Backup.BackupEnable)

	backupTab := container.NewVBox(
		widget.NewLabel("备份文件夹"),
		widget.NewEntryWithData(bd_backup_dir),
		widget.NewLabel("游戏存档文件夹"),
		widget.NewEntryWithData(bd_saved_dir),
		widget.NewLabel("备份间隔"),
		widget.NewEntryWithData(bd_backup_interval),
		widget.NewLabel("备份文件的最大数量"),
		widget.NewEntryWithData(binding.IntToString(bd_backup_max_count)),
		widget.NewCheckWithData("当备份文件数量超过最大数量时是否覆盖最早的备份文件", bd_backup_max_overwrite),
		widget.NewCheckWithData("在备份时压缩文件", bd_backup_compress),
		widget.NewCheckWithData("开启自动备份", bd_backup_enable),
		widget.NewButton("保存", func() {
			// 保存Backup配置
			var err error
			Cfg.Backup.BackupDir, err = bd_backup_dir.Get()
			if err != nil {
				LogError("bd_backup_dir.Get() failed: ", err)
			}
			Cfg.Backup.SavedDir, err = bd_saved_dir.Get()
			if err != nil {
				LogError("bd_saved_dir.Get() failed: ", err)
			}
			Cfg.Backup.BackupInterval, err = bd_backup_interval.Get()
			if err != nil {
				LogError("bd_backup_interval.Get() failed: ", err)
			}
			Cfg.Backup.BackupMaxCount, err = bd_backup_max_count.Get()
			if err != nil {
				LogError("bd_backup_max_count.Get() failed: ", err)
			}
			Cfg.Backup.BackupMaxOverwrite, err = bd_backup_max_overwrite.Get()
			if err != nil {
				LogError("bd_backup_max_overwrite.Get() failed: ", err)
			}
			Cfg.Backup.BackupCompress, err = bd_backup_compress.Get()
			if err != nil {
				LogError("bd_backup_compress.Get() failed: ", err)
			}
			Cfg.Backup.BackupEnable, err = bd_backup_enable.Get()
			if err != nil {
				LogError("bd_backup_enable.Get() failed: ", err)
			}
			LogDebug("BackupConfig: ", Cfg.Backup)
			SaveConfig()
		}),
	)

	// Restart Tab
	bd_restart_enable := binding.NewBool()
	bd_restart_enable.Set(Cfg.Restart.RestartEnable)
	bd_restart_condition := binding.NewString()
	bd_restart_condition.Set(Cfg.Restart.RestartCondition)
	bd_restart_buffer_time := binding.NewInt()
	bd_restart_buffer_time.Set(Cfg.Restart.RestartBufferTime)

	restartTab := container.NewVBox(
		widget.NewCheckWithData("是否开启自动重启", bd_restart_enable),
		widget.NewLabel("重启条件 详情请看config.ini文件"),
		widget.NewEntryWithData(bd_restart_condition),
		widget.NewLabel("重启缓冲时间"),
		widget.NewEntryWithData(binding.IntToString(bd_restart_buffer_time)),
		widget.NewButton("立刻重启", func() {
			//todo 重启 再三确认
			restartPalServer(false)
		}),
		widget.NewButton("保存", func() {
			// 保存Restart配置
			var err error
			Cfg.Restart.RestartEnable, err = bd_restart_enable.Get()
			if err != nil {
				LogError("bd_restart_enable.Get() failed: ", err)
			}
			Cfg.Restart.RestartCondition, err = bd_restart_condition.Get()
			if err != nil {
				LogError("bd_restart_condition.Get() failed: ", err)
			}
			Cfg.Restart.RestartBufferTime, err = bd_restart_buffer_time.Get()
			if err != nil {
				LogError("bd_restart_buffer_time.Get() failed: ", err)
			}
			LogDebug("RestartConfig: ", Cfg.Restart)
			SaveConfig()
		}),
	)

	// Console Tab
	Bd_LogEntry = binding.NewString()
	LogEntry = widget.NewMultiLineEntry()
	LogEntry.MultiLine = true
	LogEntry.Wrapping = fyne.TextWrapBreak

	bd_cmd := binding.NewString()
	tb_cmd := widget.NewEntryWithData(bd_cmd)
	tb_cmd.PlaceHolder = "请输入命令"
	sendCmd := func() {
		// 发送命令的逻辑
		cmd, err := bd_cmd.Get()
		bd_cmd.Set("")
		tb_cmd.Refresh()
		if err != nil {
			LogError("bd_cmd.Get() failed: ", err)
			return
		}
		LogInfo("发送命令: ", cmd)
		resp := ExecRconCmd(cmd)
		if resp == "" {
			LogError("发送命令失败")
			return
		}
		LogInfo("[Rcon]: ", resp)

	}

	tb_cmd.OnSubmitted = func(content string) {
		sendCmd()
	}

	button_send_cmd := widget.NewButton("发送命令", func() {
		sendCmd()
	})

	go func() {
		// 更新日志
		for {
			//LogEntry.Refresh()
			//time.Sleep(time.Millisecond * 300)
		}
	}()
	var button_start_server *widget.Button
	button_start_server = widget.NewButton("启动服务器", func() {
		button_start_server.Disable()
		go func() {
			// 启动Pal服务器
			if !startPalServer() {
				LogInfo("服务器启动失败")
				LogInfo("Pal 服务器启动失败, 请检查配置文件和服务器启动命令是否正确")
				button_start_server.Enable()
				return
			}
			GoGuard()
			//todo
		}()
	})
	button_start_backup := widget.NewButton("立刻备份", func() {
		go func() {
			backupPalServer()
		}()
	})

	logTab := container.NewBorder(nil,
		container.NewVBox(
			container.NewBorder(nil, nil, nil, button_send_cmd, tb_cmd),
			button_start_server,
			button_start_backup,
		),
		nil,
		nil,
		LogEntry)

	// 组合所有Tab到一个AppTabs中
	tabs := container.NewAppTabs(
		container.NewTabItem("服务器", serverTab),
		container.NewTabItem("自动备份", backupTab),
		container.NewTabItem("自动重启", restartTab),
		container.NewTabItem("控制台", logTab),
	)

	// 隐藏窗口
	if !Cfg.Server.Debug {
		hideConsole()
	}

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

// hideConsole 尝试隐藏控制台窗口
func hideConsole() {
	if !IsInWindows {
		return
	}
	var (
		kernel32         = syscall.NewLazyDLL("kernel32.dll")
		getConsoleWindow = kernel32.NewProc("GetConsoleWindow")
		showWindow       = syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	)
	consoleWindow, _, _ := getConsoleWindow.Call()
	if consoleWindow == 0 {
		return // 控制台窗口不存在
	}
	showWindow.Call(consoleWindow, 0) // SW_HIDE = 0
}
