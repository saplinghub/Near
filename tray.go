package main

import (
	"fmt"
	"os"
	"time"

	"github.com/getlantern/systray"
)

// TrayManager 托盘管理器
type TrayManager struct {
	app    *App
	frames [][]byte
}

// NewTrayManager 创建托盘管理器
func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{
		app:    app,
		frames: make([][]byte, 32),
	}
}

// Start 启动托盘
func (t *TrayManager) Start() {
	go systray.Run(t.onReady, t.onExit)
}

// onReady 托盘就绪回调
func (t *TrayManager) onReady() {
	// 加载初始图标
	iconData, err := os.ReadFile("build/appicon.png")
	if err != nil {
		fmt.Println("❌ 加载托盘图标失败:", err)
		return
	}

	systray.SetIcon(iconData)
	systray.SetTitle("")
	systray.SetTooltip("Near 倒计时")

	// 创建菜单
	mShow := systray.AddMenuItem("显示窗口", "显示主窗口")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出应用")

	// 监听菜单点击
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				t.app.ShowWindow()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	fmt.Println("✅ 系统托盘启动成功")
}

// onExit 托盘退出回调
func (t *TrayManager) onExit() {
	fmt.Println("👋 系统托盘退出")
}

// UpdateTitle 更新托盘标题
func (t *TrayManager) UpdateTitle(title string) {
	systray.SetTitle(title)
}

// StartAnimation 启动 CPU 风车动画
func (t *TrayManager) StartAnimation() {
	// 预加载所有32帧
	fmt.Println("🔄 开始加载风车动画帧...")
	for i := 0; i < 32; i++ {
		path := fmt.Sprintf("build/fan_frames/fan_%02d.png", i)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("❌ 加载风车帧 %d 失败: %v\n", i, err)
			return
		}
		t.frames[i] = data
	}
	fmt.Println("✅ 风车动画帧加载完成")

	// 等待托盘初始化
	time.Sleep(2 * time.Second)

	// 启动动画循环
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	frameIndex := 0
	lastCPUCheck := time.Now()
	currentFPS := 15

	for range ticker.C {
		// 每2秒更新一次 CPU 使用率
		if time.Since(lastCPUCheck) > 2*time.Second {
			cpuUsage := t.app.systemService.GetCPUUsage()

			// 根据 CPU 计算帧率：0% → 15fps，100% → 60fps
			currentFPS = 15 + int(cpuUsage*45.0/100.0)
			if currentFPS > 60 {
				currentFPS = 60
			}

			lastCPUCheck = time.Now()
		}

		// 根据帧率计算帧索引
		elapsed := time.Now().UnixMilli()
		frameInterval := int64(1000 / currentFPS)
		index := int((elapsed / frameInterval) % 32)

		// 只在帧变化时更新图标
		if index != frameIndex {
			systray.SetIcon(t.frames[index])
			frameIndex = index
		}
	}
}

// UpdateTrayTitle 更新托盘标题（显示置顶倒计时）
func (t *TrayManager) UpdateTrayTitle() {
	countdowns, err := t.app.GetCountdowns()
	if err != nil {
		return
	}

	// 查找置顶的倒计时
	for _, countdown := range countdowns {
		if countdown.Pinned {
			days := CalculateDays(countdown.Date)
			title := fmt.Sprintf("%d天", days)
			t.UpdateTitle(title)
			return
		}
	}

	// 没有置顶的，清空标题
	t.UpdateTitle("")
}
