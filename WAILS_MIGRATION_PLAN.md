# Tauri → Wails v2 迁移架构变更计划

## 概述

将 Near 倒计时从 Tauri 2.0 迁移到 Wails v2，保持所有功能和 UI 不变。

## 核心差异对比

| 维度 | Tauri 2.0 | Wails v2 | 迁移难度 |
|------|-----------|----------|---------|
| 后端语言 | Rust | Go | ⭐⭐⭐⭐ |
| 前端通信 | `invoke()` | `wails.Call()` | ⭐⭐ |
| 配置文件 | `tauri.conf.json` | `wails.json` | ⭐⭐ |
| 系统托盘 | 内置 API | 需要 `systray` 库 | ⭐⭐⭐ |
| 窗口管理 | `WebviewWindow` | `runtime.Window` | ⭐⭐ |
| 数据存储 | `tauri-plugin-store` | 自行实现（JSON 文件） | ⭐⭐ |
| 构建工具 | Cargo + npm | Go + npm | ⭐⭐ |
| 包大小 | ~10MB | ~15MB | - |
| 启动速度 | 快 | 中等 | - |

## 架构变更详细计划

### 1. 项目结构变更

**当前结构（Tauri）**：
```
electron-demo/
├── src/                    # Vue 前端
├── src-tauri/             # Rust 后端
│   ├── src/main.rs
│   ├── Cargo.toml
│   └── tauri.conf.json
└── package.json
```

**目标结构（Wails）**：
```
electron-demo/
├── frontend/              # Vue 前端（移动 src/ 到这里）
│   ├── src/
│   ├── package.json
│   └── vite.config.js
├── backend/               # Go 后端（新建）
│   ├── main.go           # 主程序
│   ├── app.go            # 应用逻辑
│   ├── countdown.go      # 倒计时服务
│   ├── ai.go             # AI 服务
│   ├── system.go         # 系统监控
│   └── tray.go           # 系统托盘
├── build/                 # 构建资源
│   ├── appicon.png
│   └── darwin/
├── wails.json            # Wails 配置
└── go.mod                # Go 依赖
```

### 2. 后端代码迁移（Rust → Go）

#### 2.1 主���序入口

**Rust (main.rs:314-406)**：
```rust
pub fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_store::Builder::new().build())
        .setup(|app| {
            // 创建托盘
            TrayIconBuilder::with_id("main-tray")
                .icon(icon_image)
                .on_tray_icon_event(|tray, event| { /* ... */ })
                .build(app)?;

            // 启���动画
            start_cpu_fan_tray(app.handle().clone());

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            get_countdowns,
            save_countdown,
            // ...
        ])
        .run(tauri::generate_context!())
}
```

**Go (main.go)**：
```go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "Near 倒计时",
        Width:  380,
        Height: 600,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 0},
        OnStartup:        app.startup,
        OnShutdown:       app.shutdown,
        Bind: []interface{}{
            app,
        },
        Mac: &mac.Options{
            TitleBar: mac.TitleBarHiddenInset(),
            About: &mac.AboutInfo{
                Title:   "Near 倒计时",
                Message: "越近越重要，越近越靠前",
            },
        },
    })

    if err != nil {
        println("Error:", err.Error())
    }
}
```

#### 2.2 倒计时服务

**Rust (main.rs:20-95)**：
```rust
#[tauri::command]
async fn get_countdowns(app: AppHandle) -> Result<Vec<serde_json::Value>, String> {
    let store = app.store("store.json").map_err(|e| e.to_string())?;
    Ok(store.get("countdowns")
        .and_then(|v| v.as_array().map(|a| a.clone()))
        .unwrap_or_default())
}

#[tauri::command]
async fn save_countdown(app: AppHandle, countdown: serde_json::Value) -> Result<(), String> {
    // ...
}
```

**Go (countdown.go)**：
```go
package main

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Countdown struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Date      string  `json:"date"`
    StartDate string  `json:"startDate"`
    IconType  string  `json:"iconType"`
    Order     int     `json:"order"`
    Pinned    bool    `json:"pinned"`
    CreatedAt string  `json:"createdAt"`
}

type CountdownService struct {
    dataPath string
}

func NewCountdownService() *CountdownService {
    homeDir, _ := os.UserHomeDir()
    dataPath := filepath.Join(homeDir, ".near", "store.json")
    return &CountdownService{dataPath: dataPath}
}

func (s *CountdownService) GetCountdowns() ([]Countdown, error) {
    data, err := os.ReadFile(s.dataPath)
    if err != nil {
        return []Countdown{}, nil
    }

    var store map[string]interface{}
    json.Unmarshal(data, &store)

    countdowns := []Countdown{}
    if list, ok := store["countdowns"].([]interface{}); ok {
        for _, item := range list {
            bytes, _ := json.Marshal(item)
            var countdown Countdown
            json.Unmarshal(bytes, &countdown)
            countdowns = append(countdowns, countdown)
        }
    }

    return countdowns, nil
}

func (s *CountdownService) SaveCountdown(countdown Countdown) error {
    // 读取现有数据
    data, _ := os.ReadFile(s.dataPath)
    var store map[string]interface{}
    json.Unmarshal(data, &store)

    // 更新列表
    list := []Countdown{}
    if existing, ok := store["countdowns"].([]interface{}); ok {
        for _, item := range existing {
            bytes, _ := json.Marshal(item)
            var c Countdown
            json.Unmarshal(bytes, &c)
            if c.ID != countdown.ID {
                list = append(list, c)
            }
        }
    }
    list = append(list, countdown)

    store["countdowns"] = list

    // 保存
    bytes, _ := json.MarshalIndent(store, "", "  ")
    os.MkdirAll(filepath.Dir(s.dataPath), 0755)
    return os.WriteFile(s.dataPath, bytes, 0644)
}
```

#### 2.3 系统托盘

**Rust (main.rs:318-368)**：
```rust
TrayIconBuilder::with_id("main-tray")
    .icon(icon_image)
    .on_tray_icon_event(move |_tray, event| {
        match event {
            TrayIconEvent::Click { button: MouseButton::Left, .. } => {
                // 显示/隐藏窗口
            }
            _ => {}
        }
    })
    .build(app)?;
```

**Go (tray.go)** - 使用 `github.com/getlantern/systray`：
```go
package main

import (
    "github.com/getlantern/systray"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type TrayManager struct {
    app *App
}

func NewTrayManager(app *App) *TrayManager {
    return &TrayManager{app: app}
}

func (t *TrayManager) Start() {
    go systray.Run(t.onReady, t.onExit)
}

func (t *TrayManager) onReady() {
    // 设置图标
    iconData, _ := os.ReadFile("build/appicon.png")
    systray.SetIcon(iconData)
    systray.SetTitle("")
    systray.SetTooltip("Near 倒计时")

    // 创建菜单
    mQuit := systray.AddMenuItem("退出", "退出应用")

    // 监听点击事件
    go func() {
        for {
            select {
            case <-mQuit.ClickedCh:
                systray.Quit()
            }
        }
    }()

    // 监听托盘图标点击（需要额外处理）
    // Wails v2 不直接支持托盘图标点击，需要使用 systray 库
}

func (t *TrayManager) onExit() {
    // 清理
}

func (t *TrayManager) UpdateTitle(title string) {
    systray.SetTitle(title)
}
```

**⚠️ 重大问题**：
- Wails v2 **不原生支持系统托盘点击显示窗口**
- 需要使用第三方库 `systray`，但功能有限
- **无法获取托盘图标的屏幕坐标**，多显示器定位会更困难

#### 2.4 系统监控

**Rust (main.rs:97-138)** - 使用 `sysinfo`：
```rust
#[tauri::command]
async fn get_system_stats() -> Result<serde_json::Value, String> {
    let mut sys = System::new_all();
    sys.refresh_all();

    let cpu_usage = sys.global_cpu_info().cpu_usage();
    let total_memory = sys.total_memory();
    let used_memory = sys.used_memory();

    Ok(json!({
        "cpu": cpu_usage,
        "memory": memory_usage,
        "temperature": temperature,
        "total_memory": total_memory,
        "used_memory": used_memory
    }))
}
```

**Go (system.go)** - 使用 `github.com/shirou/gopsutil`：
```go
package main

import (
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
)

type SystemStats struct {
    CPU         float64 `json:"cpu"`
    Memory      float64 `json:"memory"`
    Temperature string  `json:"temperature"`
    TotalMemory uint64  `json:"total_memory"`
    UsedMemory  uint64  `json:"used_memory"`
}

func (a *App) GetSystemStats() (*SystemStats, error) {
    // CPU
    cpuPercent, _ := cpu.Percent(0, false)
    cpuUsage := 0.0
    if len(cpuPercent) > 0 {
        cpuUsage = cpuPercent[0]
    }

    // 内存
    vmStat, _ := mem.VirtualMemory()

    // 温度（macOS）
    temperature := "N/A"
    // Go 获取 macOS 温度较困难，需要调用 ioreg 命令

    return &SystemStats{
        CPU:         cpuUsage,
        Memory:      vmStat.UsedPercent,
        Temperature: temperature,
        TotalMemory: vmStat.Total,
        UsedMemory:  vmStat.Used,
    }, nil
}
```

#### 2.5 CPU 风车动画

**Rust (main.rs:244-312)**：
```rust
fn start_cpu_fan_tray(app: AppHandle) {
    // 预加载32帧
    let frames: Vec<Image> = (0..32).map(|i| {
        let path = format!("icons/fan_frames/fan_{:02}.png", i);
        // 加载图片
    }).collect();

    thread::spawn(move || {
        loop {
            let cpu_usage = sys.global_cpu_info().cpu_usage();
            let current_fps = 15 + ((cpu_usage * 45.0 / 100.0) as u32).min(60);

            // 更新托盘图标
            tray_ref.set_icon(Some(frames[index].clone()));

            std::thread::sleep(Duration::from_millis(50));
        }
    });
}
```

**Go (tray.go)**：
```go
func (t *TrayManager) StartAnimation() {
    // 预加载32帧
    frames := make([][]byte, 32)
    for i := 0; i < 32; i++ {
        data, _ := os.ReadFile(fmt.Sprintf("build/fan_frames/fan_%02d.png", i))
        frames[i] = data
    }

    go func() {
        ticker := time.NewTicker(50 * time.Millisecond)
        defer ticker.Stop()

        frameIndex := 0
        for range ticker.C {
            // 获取 CPU 使用率
            cpuPercent, _ := cpu.Percent(0, false)
            cpuUsage := 0.0
            if len(cpuPercent) > 0 {
                cpuUsage = cpuPercent[0]
            }

            // 计算帧率
            fps := 15 + int(cpuUsage*45.0/100.0)
            if fps > 60 {
                fps = 60
            }

            // 更新图标
            frameIndex = (frameIndex + 1) % 32
            systray.SetIcon(frames[frameIndex])
        }
    }()
}
```

**⚠️ 问题**：
- `systray` 库的 `SetIcon()` 性能不如 Tauri
- 可能导致动画不够流畅

### 3. 前端代码变更

#### 3.1 API 调用方式

**Tauri (App.vue)**：
```javascript
import { invoke } from '@tauri-apps/api/core';

// 获取倒计时列表
const countdowns = await invoke('get_countdowns');

// 保存倒计时
await invoke('save_countdown', { countdown: data });
```

**Wails (App.vue)**：
```javascript
import { GetCountdowns, SaveCountdown } from '../wailsjs/go/main/App';

// 获取倒计时列表
const countdowns = await GetCountdowns();

// 保存倒计时
await SaveCountdown(data);
```

**变更点**：
- 移除 `@tauri-apps/api` 依赖
- 使用 Wails 自动生成的 TypeScript 绑定
- 函数名从 `snake_case` 改为 `PascalCase`

#### 3.2 窗口管理

**Tauri**：
```javascript
import { getCurrentWindow } from '@tauri-apps/api/window';

const win = getCurrentWindow();
win.hide();
win.show();
```

**Wails**：
```javascript
import { WindowHide, WindowShow } from '../wailsjs/runtime/runtime';

WindowHide();
WindowShow();
```

#### 3.3 构建配置

**Tauri (vite.config.js)**：
```javascript
export default defineConfig({
  plugins: [vue()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
  },
  envPrefix: ['VITE_', 'TAURI_'],
});
```

**Wails (frontend/vite.config.js)**：
```javascript
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 34115, // Wails 默认端口
  },
  build: {
    outDir: 'dist',
  },
});
```

### 4. 配置文件变更

#### 4.1 Tauri 配置

**tauri.conf.json**：
```json
{
  "productName": "Near",
  "version": "1.0.0",
  "identifier": "com.near.countdown",
  "build": {
    "beforeDevCommand": "npm run dev",
    "beforeBuildCommand": "npm run build",
    "devUrl": "http://localhost:5173",
    "frontendDist": "../dist"
  },
  "app": {
    "windows": [{
      "title": "Near 倒计时",
      "width": 380,
      "height": 600,
      "decorations": false,
      "transparent": true
    }],
    "security": {
      "csp": null
    }
  }
}
```

#### 4.2 Wails 配置

**wails.json**：
```json
{
  "name": "Near",
  "outputfilename": "Near",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Sapling",
    "email": "your@email.com"
  },
  "info": {
    "companyName": "Near",
    "productName": "Near 倒计时",
    "productVersion": "1.0.0",
    "copyright": "Copyright © 2024",
    "comments": "越近越重要，越近越靠前"
  }
}
```

### 5. 依赖变更

#### 5.1 Go 依赖 (go.mod)

```go
module near

go 1.21

require (
    github.com/wailsapp/wails/v2 v2.8.0
    github.com/getlantern/systray v1.2.2
    github.com/shirou/gopsutil/v3 v3.23.12
)
```

#### 5.2 前端依赖 (package.json)

**移除**：
```json
{
  "@tauri-apps/api": "^2.0.0",
  "@tauri-apps/cli": "^2.0.0"
}
```

**保留**：
```json
{
  "vue": "^3.4.0",
  "vite": "^5.4.0",
  "sortablejs": "^1.15.0"
}
```

### 6. 构建流程变更

#### 6.1 开发模式

**Tauri**：
```bash
npm run tauri dev
```

**Wails**：
```bash
wails dev
```

#### 6.2 生产构建

**Tauri**：
```bash
npm run tauri build
```

**Wails**：
```bash
wails build
```

### 7. 功能对比与风险评估

| 功能 | Tauri 实现 | Wails 实现 | 风险等级 |
|------|-----------|-----------|---------|
| 倒计时管理 | ✅ 完整 | ✅ 可实现 | 🟢 低 |
| 拖拽排序 | ✅ 前端实现 | ✅ 前端实现 | 🟢 低 |
| AI 解析 | ✅ HTTP 调用 | ✅ HTTP 调用 | 🟢 低 |
| 数据存储 | ✅ tauri-plugin-store | ⚠️ 手动实现 JSON | 🟡 中 |
| 系统监控 | ✅ sysinfo | ✅ gopsutil | 🟢 低 |
| 系统托盘 | ✅ 原生支持 | ⚠️ 第三方库 | 🔴 高 |
| 托盘点击定位 | ⚠️ 有 bug | 🔴 更困难 | 🔴 高 |
| CPU 风车动画 | ✅ 流畅 | ⚠️ 可能卡顿 | 🟡 中 |
| 失焦隐藏 | ✅ 原生支持 | ✅ 可实现 | 🟢 低 |
| 多显示器支持 | 🔴 有 bug | 🔴 更困难 | 🔴 高 |

### 8. 迁移工作量估算

| 任务 | 工作量 | 优先级 |
|------|--------|--------|
| 项目结构调整 | 2小时 | P0 |
| Go 后端基础框架 | 4小时 | P0 |
| 倒计时服务迁移 | 3小时 | P0 |
| 系统监控迁移 | 2小时 | P1 |
| 系统托盘实现 | 6小时 | P0 |
| CPU 风车动画 | 4小时 | P1 |
| 前端 API 调用改造 | 3小时 | P0 |
| 数据存储实现 | 2小时 | P1 |
| 多显示器定位 | 8小时 | P2 |
| 测试与调试 | 6小时 | P0 |
| **总计** | **40小时** | - |

### 9. 关键风险与建议

#### 🔴 高风险项

1. **系统托盘功能受限**
   - Wails v2 不原生支持托盘图标点击事件
   - 第三方库 `systray` 功能有限
   - **无法获取托盘图标屏幕坐标**，多显示器定位几乎不可能实现

2. **多显示器支持更困难**
   - Tauri 已经有 bug，Wails 会更糟
   - 建议：放弃多显示器精确定位，固定显示在主显示器

3. **CPU 风车动画性能**
   - `systray.SetIcon()` 性能不如 Tauri
   - 可能需要降低帧率或简化动画

#### 🟡 中风险项

1. **数据存储需要手动实现**
   - 需要自己处理 JSON 文件读写
   - 需要处理并发访问和文件锁

2. **Go 学习曲线**
   - 如果不熟悉 Go，需要学习时间
   - Go 的错误处理方式与 Rust 不同

#### 🟢 低风险项

1. **前端代码改动小**
   - 主要是 API 调用方式变更
   - UI/样式完全不变

2. **基础功能可实现**
   - 倒计时管理、AI 解析等核心功能都能实现

### 10. 最终建议

#### ❌ 不建议迁移的理由

1. **系统托盘功能严重受限**
   - 这是本应用的核心交互方式
   - Wails v2 的托盘支持远不如 Tauri

2. **多显示器问题会更严重**
   - Tauri 的问题在 Wails 中无法解决
   - 甚至可能完全无法实现

3. **性能可能下降**
   - CPU 风车动画可能不够流畅
   - Go 的 GC 可能导致卡顿

4. **工作量大，收益小**
   - 需要 40+ 小时重写后端
   - 功能不增反减

#### ✅ 如果坚持迁移，建议

1. **简化功能**
   - 放弃 CPU 风车动画
   - 放弃多显示器精确定位
   - 简化为普通窗口应用（不依赖托盘）

2. **改变交互方式**
   - 不使用托盘图标点击
   - 改为菜单栏应用或 Dock 应用
   - 使用全局快捷键唤醒

3. **分阶段迁移**
   - 先实现基础功能（倒计时管理）
   - 再逐步添加高级功能
   - 最后处理托盘和动画

## 总结

**Tauri → Wails v2 迁移不推荐**，主要原因：

1. 🔴 系统托盘功能严重受限
2. 🔴 多显示器支持更困难
3. 🟡 性能可能下降
4. 🟡 工作量大（40+ 小时）
5. 🟢 功能不增反减

**建议**：
- 保持 Tauri 2.0，等待 Tauri 2.6+ 解决多显示器问题
- 或者考虑 Electron（虽然包体积大，但功能最完整）
- 如果必须用 Go，考虑 Fyne 或 Qt for Go（但都不如 Tauri）

---

**是否继续迁移？请确认后我再开始实施。**
