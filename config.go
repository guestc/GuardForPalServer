package main

import (
	"gopkg.in/ini.v1"
)

// ServerConfig 服务器设置
type ServerConfig struct {
	ExecStart string `ini:"execstart"`     // 服务器启动路径 支持带参数
	RconIP    string `ini:"rcon_ip"`       // rcon IP地址 默认为本地
	RconPort  int    `ini:"rcon_port"`     // rcon 端口 默认为25575
	RconPass  string `ini:"rcon_password"` // rcon 密码
	Debug     bool   `ini:"debug"`         // 是否开启调试模式
}

// BackupConfig 备份设置
type BackupConfig struct {
	BackupDir          string `ini:"backup_dir"`           // 存放备份文件的文件夹
	SavedDir           string `ini:"saved_dir"`            // 游戏存档文件夹
	BackupInterval     string `ini:"backup_interval"`      // 备份间隔
	BackupMaxCount     int    `ini:"backup_max_count"`     // 备份文件的最大数量 0为不限制
	BackupMaxOverwrite bool   `ini:"backup_max_overwrite"` // 当备份文件数量超过最大数量时是否覆盖最早的备份文件
	BackupCompress     bool   `ini:"backup_compress"`      // 是否在备份时压缩文件
	BackupEnable       bool   `ini:"backup_enable"`        // 开启自动备份
}

// RestartConfig 自动重启设置
type RestartConfig struct {
	RestartEnable     bool   `ini:"restart_enable"`      // 是否开启自动重启
	RestartCondition  string `ini:"restart_condition"`   // 重启条件
	RestartBufferTime int    `ini:"restart_buffer_time"` // 重启缓冲时间
}

// Config 总配置
type AppConfig struct {
	Server  ServerConfig  `ini:"server"`  // 服务器设置
	Backup  BackupConfig  `ini:"backup"`  // 备份设置
	Restart RestartConfig `ini:"restart"` // 自动重启设置
}

var Cfg AppConfig
var CfgPath string = "./config.ini"

// 初始化 Config 配置
func InitConfig() bool {
	err := ini.MapTo(&Cfg, CfgPath)
	if err != nil {
		LogError("读取配置文件失败: " + err.Error())
		LogError("请检查配置文件是否存在并且格式正确")
		return false
	}
	LogInfo("读取配置文件成功")
	DumpConfig()
	return true
}

// 保存配置文件
func SaveConfig() {
	cfg_file, err := ini.Load(CfgPath)
	if err != nil {
		LogError("Fail to read file: %v", err)
		return
	}
	err = cfg_file.ReflectFrom(&Cfg)
	if err != nil {
		LogError("Fail to Reflect: %v", err)
		return
	}
	err = cfg_file.SaveTo(CfgPath)
	if err != nil {
		LogError("Fail to save file: %v", err)
		return
	}
	LogInfo("保存配置文件成功")
}

func DumpConfig() {
	LogDebug("ServerConfig: ", Cfg.Server)
	LogDebug("ExecStart: ", Cfg.Server.ExecStart)
	LogDebug("RconIP: ", Cfg.Server.RconIP)
	LogDebug("RconPort: ", Cfg.Server.RconPort)
	LogDebug("RconPass: ", Cfg.Server.RconPass)
	LogDebug("Debug: ", Cfg.Server.Debug)

	LogDebug("BackupConfig: ", Cfg.Backup)
	LogDebug("BackupDir: ", Cfg.Backup.BackupDir)
	LogDebug("SavedDir: ", Cfg.Backup.SavedDir)
	LogDebug("BackupInterval: ", Cfg.Backup.BackupInterval)
	LogDebug("BackupMaxCount: ", Cfg.Backup.BackupMaxCount)
	LogDebug("BackupMaxOverwrite: ", Cfg.Backup.BackupMaxOverwrite)
	LogDebug("BackupCompress: ", Cfg.Backup.BackupCompress)
	LogDebug("BackupEnable: ", Cfg.Backup.BackupEnable)

	LogDebug("RestartConfig: ", Cfg.Restart)
	LogDebug("RestartEnable: ", Cfg.Restart.RestartEnable)
	LogDebug("RestartCondition: ", Cfg.Restart.RestartCondition)
	LogDebug("RestartBufferTime: ", Cfg.Restart.RestartBufferTime)
}
