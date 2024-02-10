package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorcon/rcon"
)

var PalServerPid int
var ReadPalServerPid int32
var Rcon_conn *rcon.Conn
var IsInWindows bool
var AppVserion string = "1.0.5"

func test() {

}

func main() {
	IsInWindows = runtime.GOOS == "windows"
	// 初始化配置
	if !InitConfig() {
		LogInfo("回车退出：")
		fmt.Scan()
		os.Exit(1)
	}
	main_ui()
}

// 启动Pal服务器
func startPalServer() bool {
	// 执行启动命令
	sargs := strings.Split(Cfg.Server.ExecStart, " ")
	exitCode := -1

	//进入协程
	go func(exitCode *int) {
		cmd := exec.Command(sargs[0], sargs[1:]...)
		// 获取命令的标准输出流
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			LogError("Error creating StdoutPipe for Cmd", err)
			*exitCode = 1
			return
		}
		//不好获取标准输出流 干脆不弄了
		defer stdoutPipe.Close()

		// 启动命令
		if err := cmd.Start(); err != nil {
			LogError("Error starting Cmd", err)
			*exitCode = 1
			return
		}
		//更Pal服务器PID
		PalServerPid = cmd.Process.Pid
		LogInfo("Pal服务器 PID: ", cmd.Process.Pid)
		*exitCode = 0
		// 等待命令执行完成
		if err := cmd.Wait(); err != nil {
			LogError("Error waiting for Cmd", err)
			*exitCode = 1
			return
		}
	}(&exitCode)
	//todo 获取标准输出流了再说
	for {
		if exitCode != -1 {
			LogInfo("startPalServer : ", exitCode)
			if exitCode == 0 {
				return connectRcon()
			} else {
				return false
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// 开始链接Rcon服务器
func connectRcon() bool {
	time.Sleep(6 * time.Second)
	var err error
	for x := 0; x < 10; x++ {
		LogInfo("正在连接Rcon服务器, 尝试次数: ", x+1)
		Rcon_conn, err = rcon.Dial(fmt.Sprintf("%s:%d", Cfg.Server.RconIP, Cfg.Server.RconPort), Cfg.Server.RconPass)
		if err == nil {
			break
		}
		LogDebug("连接Rcon服务器失败: ", err)
		time.Sleep(2 * time.Second)
	}
	defer Rcon_conn.Close()
	if err != nil {
		LogError("连接Rcon服务器失败: ", err)
		LogError("请检查Rcon服务器IP, 端口和密码是否正确")
		LogError("Rcon服务器IP: ", Cfg.Server.RconIP)
		LogError("Rcon服务器端口: ", Cfg.Server.RconPort)
		LogError("Rcon服务器密码: ", Cfg.Server.RconPass)
		return false
	}

	resp, err := Rcon_conn.Execute("info")
	if err != nil {
		LogError("发送命令失败: ", err)
		return false
	}

	LogInfo("Rcon服务器返回: ", resp)
	if strings.Contains(resp, "Welcome to Pal Server") {
		LogInfo("Rcon服务器连接成功")
		showServerCmdList()
		ReadPalServerPid, err = getPalServerProcess(int32(PalServerPid))
		if err != nil {
			LogError("获取Pal服务器进程失败: ", err)
			return false
		}
		LogDebug("Pal服务器进程: ", ReadPalServerPid)
		return true
	} else {
		LogError("Rcon服务器返回错误")
		return false
	}
}

// 守护进程
func GoGuard() {
	//todo
	LogInfo("守护进程启动")
	var backup_timespan = time.Now()
	var restart_timespan = time.Now()
	go func() {
		for {

			//判断 服务端卡死 崩溃
			if getProcessMemoryUsage(ReadPalServerPid) == 0 && ExecRconCmd("info") == "" {
				//todo
				LogError("Pal服务器崩溃")
				LogInfo("Pal服务器崩溃, 重启服务器...")
				killPid(ReadPalServerPid)
				killPid(int32(PalServerPid))
				time.Sleep(5 * time.Second)
				restartPalServer(true)
			}

			// 自动备份
			if Cfg.Backup.BackupEnable {
				// 备份间隔
				need_backup, err := meetTimeConditins(backup_timespan, Cfg.Backup.BackupInterval)
				if err != nil {
					LogError("Backup->meetTimeConditins failed: ", err)
				} else {
					if need_backup {
						backupPalServer()
						backup_timespan = time.Now()
					}
				}

			}

			// 自动重启
			if Cfg.Restart.RestartEnable {
				if reMatch(`\d+[GM]`, Cfg.Restart.RestartCondition) {
					//按内存大小重启
					var size = reFind(`(\d+)[GM]`, Cfg.Restart.RestartCondition)
					size = reFind(`(\d+)`, size)
					c_bb := getProcessMemoryUsage(ReadPalServerPid)
					if c_bb == 0 {
						LogError("获取Pal服务器内存使用失败")
					} else {
						c_mb, c_gb := byte2mb(c_bb), byte2gb(c_bb)
						need_restart := false
						if strings.Contains(Cfg.Restart.RestartCondition, "G") {
							g, err := strconv.ParseUint(size, 10, 64)
							if err != nil {
								LogError("strconv.ParseUint failed: ", err)
							} else {
								if c_gb >= g {
									need_restart = true
								}
							}
						}
						if strings.Contains(Cfg.Restart.RestartCondition, "M") {
							m, err := strconv.ParseUint(size, 10, 64)
							if err != nil {
								LogError("strconv.ParseUint failed: ", err)
							} else {
								if c_mb >= m {
									need_restart = true
								}
							}
						}
						if need_restart {
							//todo 需要重启
							restartPalServer(false)
							restart_timespan = time.Now()
						}
					}

				}
				if reMatch(`\d+[dhm]`, Cfg.Restart.RestartCondition) {
					// 按时间重启
					need_restart, err := meetTimeConditins(restart_timespan, reFind(`\d+[dhm]`, Cfg.Restart.RestartCondition))
					if err != nil {
						LogError("Restart->meetTimeConditins failed: ", err)
					} else {
						if need_restart {

							restartPalServer(false)
							restart_timespan = time.Now()
							//todo
						}
					}
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func isProcessRunning(i int32) {
	panic("unimplemented")
}

// 重启服务器
func restartPalServer(is_crash bool) {
	LogInfo(">开始重启服务器......")
	if !is_crash {
		wait_time := Cfg.Restart.RestartBufferTime
		for ; wait_time > 0; wait_time-- {
			if wait_time%10 == 0 {
				ExecRconCmd(fmt.Sprintf("Broadcast Server will be restart in %d s", wait_time))
			}
			time.Sleep(1 * time.Second)
		}
		ExecRconCmd("Save")
		time.Sleep(5 * time.Second)
		ExecRconCmd("DoExit")
		backupPalServer()
	}
	time.Sleep(5 * time.Second)
	killPid(ReadPalServerPid)
	killPid(int32(PalServerPid))
	time.Sleep(5 * time.Second)
	// 启动Pal服务器
	for {
		if !startPalServer() {
			LogInfo("服务器启动失败")
			LogInfo("Pal 服务器启动失败, 请检查配置文件和服务器启动命令是否正确")
			LogInfo("30秒后重试...")
			time.Sleep(30 * time.Second)
		} else {
			break
		}
	}
}

// 备份服务器游戏存档
func backupPalServer() bool {
	LogInfo(">>>>>>>>>>>>>>开始备份")
	if !Cfg.Backup.BackupEnable {
		return false
	}
	if Cfg.Backup.BackupMaxOverwrite {
		if Cfg.Backup.BackupMaxCount > 0 {
			// 检测备份文件数量是不是超了
			if getBackupCount() >= Cfg.Backup.BackupMaxCount {
				//超了
				//删除最早的备份
				LogInfo("备份文件数量超过最大数量, 删除最早的备份")
				deleteEarliestBackup()
			}
		}
	}
	//保存存档
	ExecRconCmd("Save")
	time.Sleep(5 * time.Second)
	root := Cfg.Backup.BackupDir
	day_dir := fmt.Sprintf("%s/%s", root, strings.Replace(time.Now().Format("2006-01-02"), ":", "", -1))

	_, err := os.Stat(day_dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(day_dir, os.ModePerm)
			if err != nil {
				return false
			}
		}
	}

	//保存存档
	ExecRconCmd("Save")
	time.Sleep(5 * time.Second)

	// 备份为ZIP
	if Cfg.Backup.BackupCompress {
		zip_file := fmt.Sprintf("%s/%s.zip", day_dir, time.Now().Format("15h04m05s"))
		err := zipFile(Cfg.Backup.SavedDir, zip_file)
		if err != nil {
			LogError("压缩备份失败: ", err)
			return false
		}
		LogInfo("备份成功: ", zip_file)
		return true
	} else {
		// 复制存档
		cp_dir := fmt.Sprintf("%s/%s", day_dir, time.Now().Format("15h04m05s"))
		os.MkdirAll(cp_dir, os.ModePerm)
		err := copyDir(Cfg.Backup.SavedDir, cp_dir)
		if err != nil {
			LogError("复制备份失败: ", err)
			return false
		}
		LogInfo("备份成功: ", cp_dir)
		return true
	}
}
