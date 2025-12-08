package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用结构体
type App struct {
	ctx              context.Context
	countdownService *CountdownService
	systemService    *SystemService
	trayManager      *TrayManager
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{
		countdownService: NewCountdownService(),
		systemService:    NewSystemService(),
	}
}

// startup 在应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化托盘管理器
	a.trayManager = NewTrayManager(a)
	a.trayManager.Start()

	// 启动 CPU 风车动画
	go a.trayManager.StartAnimation()

	// 启动定时更新托盘标题
	go a.startTrayTitleUpdater()

	fmt.Println("🚀 Near 倒计时启动成功")
}

// shutdown 在应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("👋 Near 倒计时关闭")
}

// startTrayTitleUpdater 定时更新托盘标题
func (a *App) startTrayTitleUpdater() {
	// TODO: 实现定时更新逻辑
}

// 窗口管理方法

// ShowWindow 显示窗口
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

// HideWindow 隐藏窗口
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// ToggleWindow 切换窗口显示/隐藏
func (a *App) ToggleWindow() {
	// Wails v2 没有直接的 IsVisible 方法，需要自己维护状态
	// 这里简化处理，直接显示
	runtime.WindowShow(a.ctx)
}
