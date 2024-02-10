package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorcon/rcon"
	"github.com/otiai10/copy"
	"github.com/shirou/gopsutil/process"
)

// 执行Rcon命令
func ExecRconCmd(cmd string) string {
	if cmd == "help" {
		showServerCmdList()
		return ""

	}
	var err error
	Rcon_conn, err = rcon.Dial(fmt.Sprintf("%s:%d", Cfg.Server.RconIP, Cfg.Server.RconPort), Cfg.Server.RconPass)
	if err != nil {
		LogError("Rcon connection failed: ", err)
		return ""
	}
	defer Rcon_conn.Close()
	resp, err := Rcon_conn.Execute(cmd)
	if err != nil {
		LogError("Rcon command failed: ", err)
		return ""
	}
	return strings.ReplaceAll(strings.ReplaceAll(resp, "\r\n", "\n"), "\n\n", "\n")
}

// parseInterval 将备份间隔字符串转换为总秒数
func parseInterval(interval string) (int, error) {
	var totalSeconds int
	// 使用正则表达式匹配天、小时和分钟
	re := regexp.MustCompile(`(\d+)([dhm])`)
	matches := re.FindAllStringSubmatch(interval, -1)

	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}

		switch match[2] {
		case "d":
			totalSeconds += value * 86400 // 1天=86400秒
		case "h":
			totalSeconds += value * 3600 // 1小时=3600秒
		case "m":
			totalSeconds += value * 60 // 1分钟=60秒
		}
	}
	return totalSeconds, nil
}

// meetTimeConditins 判断是否满足时间条件
func meetTimeConditins(last time.Time, interval string) (bool, error) {
	intervalSeconds, err := parseInterval(interval)
	if err != nil {
		return false, err
	}
	// 计算自上次以来经过的时间（秒）
	elapsed := time.Since(last).Seconds()

	return elapsed >= float64(intervalSeconds), nil
}

func byte2mb(b uint64) uint64 {
	return b / 1024 / 1024
}
func byte2gb(b uint64) uint64 {
	return b / 1024 / 1024 / 1024
}

// 通过父进程ID获取Pal服务器进程
func getPalServerProcess(parentPid int32) (int32, error) {
	// 获取系统中的所有进程
	processes, err := process.Processes()
	if err != nil {
		return 0, err
	}

	for _, proc := range processes {
		ppid, err := proc.Ppid() // 获取进程的父进程ID
		if err != nil {
			continue // 如果无法获取PPID，跳过此进程
		}
		name, err := proc.Name()
		if err != nil {
			continue
		}

		// 如果父进程ID匹配，将进程添加到子进程列表中
		if ppid == parentPid {
			//LogDebug("Found child process: ", name, ppid)
			if strings.Contains(name, "PalServer") {
				return proc.Pid, nil
			}
		}
	}
	return 0, nil
}

// 获取进程的内存使用情况
func getProcessMemoryUsage(pid int32) uint64 {
	proc, err := process.NewProcess(pid)
	if err != nil {
		fmt.Printf("Failed to create process instance: %s\n", err)
		return 0
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		fmt.Printf("Failed to get memory info: %s\n", err)
		return 0
	}

	// fmt.Printf("Memory Usage of PID %d:\n", pid)
	// fmt.Printf("RSS: %v bytes\n", memInfo.RSS) // 常驻集大小
	// fmt.Printf("VMS: %v bytes\n", memInfo.VMS) // 虚拟内存大小
	return memInfo.RSS
}

// kill进程
func killPid(pid int32) error {
	if IsInWindows {
		cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	} else {
		// 在Unix-like系统上，尝试发送SIGINT
		process, err := os.FindProcess(int(pid))
		if err != nil {
			return err
		}
		return process.Signal(os.Interrupt) // 或使用syscall.SIGKILL等
	}
}

// 复制文件夹
func copyDir(src, dst string) error {
	return copy.Copy(src, dst)
}

// 正则检测
func reMatch(pattern, str string) bool {
	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return false
	}
	match := re.MatchString(str)
	return match
}

// 正则搜索
func reFind(pattern, str string) string {
	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return ""
	}
	match := re.FindString(str)
	return match
}

// 获取备份数量
func getBackupCount() int {
	root := Cfg.Backup.BackupDir
	count := 0
	visit := func(files *[]string) filepath.WalkFunc {
		return func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println("error:", err)
				return err
			}
			*files = append(*files, path)
			return nil
		}
	}
	var files []string
	err := filepath.Walk(root, visit(&files))
	if err != nil {
		LogWarning("获取备份数量失败: ", err)
		return count
	}
	if Cfg.Backup.BackupCompress {
		for _, file := range files {
			// LogDebug("file: ", file)
			if reMatch(`\d{2}h\d{2}m\d{2}s.zip`, file) {
				count++
			}
		}
	} else {
		for _, file := range files {
			// LogDebug("file: ", file)
			if reMatch(`\d{2}h\d{2}m\d{2}s`, file) {
				count++
			}
		}
	}
	return count

}

type backupFile struct {
	name string
	time int64
	dir  bool
}

// 删除最早的备份文件
func deleteEarliestBackup() error {
	root := Cfg.Backup.BackupDir
	visit := func(files *[]string) filepath.WalkFunc {
		return func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println("error:", err)
				return err
			}
			*files = append(*files, path)
			return nil
		}
	}
	var files []string
	err := filepath.Walk(root, visit(&files))
	if err != nil {
		LogError("获取备份数据失败: ", err)
		return err
	}

	var matchedFiles []backupFile
	if Cfg.Backup.BackupCompress {
		for _, file := range files {
			// LogDebug("file: ", file)
			if reMatch(`\d{2}h\d{2}m\d{2}s.zip`, file) {
				fileInfo, err := os.Stat(file)
				if err != nil {
					LogError("获取文件信息失败: ", err)
					continue
				}
				matchedFiles = append(matchedFiles, backupFile{name: file, time: fileInfo.ModTime().Unix(), dir: false})
			}
		}
	} else {
		for _, file := range files {
			// LogDebug("file: ", file)
			if reMatch(`\d{2}h\d{2}m\d{2}s`, file) {
				fileInfo, err := os.Stat(file)
				if err != nil {
					LogError("获取文件信息失败: ", err)
					continue
				}
				matchedFiles = append(matchedFiles, backupFile{name: file, time: fileInfo.ModTime().Unix(), dir: true})
			}
		}
	}
	var earliest backupFile
	for _, mfile := range matchedFiles {
		if earliest == (backupFile{}) {
			earliest = mfile
		} else {
			if mfile.time < earliest.time {
				earliest = mfile
			}
		}
	}
	LogDebug("earliest: ", earliest)
	if earliest != (backupFile{}) {
		if earliest.dir {
			os.RemoveAll(earliest.name)
		} else {
			os.Remove(earliest.name)
		}
	}
	return nil
}

// 显示服务器指令列表
func showServerCmdList() {
	LogInfo("服务器指令列表: ( 在控制台不需要加/ )")
	LogInfo("/Shutdown {秒} {信息} 向服务器发送信息然后在指定时间关服例如：/Shutdown 10 10秒后即将关服")
	LogInfo("/DoExit 直接关服")
	LogInfo("/KickPlayer {steamID} 踢出指定玩家，玩家的steamid可以通过/ShowPlayers查看")
	LogInfo("/BanPlayer {SteamID} 禁止指定玩家进服")
	LogInfo("/TeleportToPlayer {SteamID} 把自己传送到指定玩家")
	LogInfo("/TeleportToMe {SteamID} 把指定的玩家传送到自己身边")
	LogInfo("/ShowPlayers 查看服务器在线玩家信息")
	LogInfo("/Save 保存服务器存档")
	LogInfo("/help 显示帮助信息")
	if !IsInWindows {
		LogInfo("/restart 重启服务器	(Linux端专用)")
		LogInfo("/backup 立刻备份	(Linux端专用)")
		LogInfo("/stop 关闭服务器并且退出守护程序	(Linux端专用)")
		LogInfo("/reloadcfg 重新加载配置文件	(Linux端专用)")
	}

}
