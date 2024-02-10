@echo off
chcp 65001
setlocal enabledelayedexpansion rem 开启延迟变量
set /p platform=输入平台Windows(w) 或者 Linux(l)：
set BUILD_FILE=.
if %platform%==w (
    set GOOS=windows
    set GOARCH=amd64
    set OUTPUT_FILE=.\output\windows\GuardForPalServer.exe
    echo "Building for windows"
    echo !OUTPUT_FILE!
    go build -o !OUTPUT_FILE! %BUILD_FILE%
    echo "Build complete"
    copy /Y .\config.ini .\output\windows
    exit
)
if %platform%==l (
    set GOOS=linux
    set GOARCH=amd64
    set OUTPUT_FILE=.\output\linux\GuardForPalServer
    echo "Building for linux"
    echo !OUTPUT_FILE!
    go build -o !OUTPUT_FILE! %BUILD_FILE%
    echo "Build complete"
    copy /Y .\config.ini .\output\linux
    exit
)