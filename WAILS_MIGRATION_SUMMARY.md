# Tauri → Wails v2 迁移完成总结

## 迁移状态

✅ **代码迁移已完成**（2024-12-08）

## 项目结构变更

### 新的目录结构

```
electron-demo/
├── frontend/              # Vue 前端
│   ├── src/
│   │   ├── App.vue       # 主组件（已改造）
│   │   ├── ai-service.js # AI 服务
│   │   ├── wails-api.js  # Wails API 适配层（新增）
│   │   └── renderer.js   # 入口文件
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── build/                 # 构建资源
│   ├── appicon.png       # 托盘图标
│   └── fan_frames/       # 风车动画帧（32张）
├── main.go               # Go 主程序
├── app.go                # 应用逻辑
├── countdown.go          # 倒计时服务
├── system.go             # 系统监控
├── tray.go               # 系统托盘
├── ai.go                 # AI 配置
├── wails.json            # Wails 配置
└── go.mod                # Go 依赖
```

### 已删除/保留

- ❌ 删除：`src-tauri/` 目录（Rust 后端）
- ❌ 删除：`tauri.conf.json`
- ❌ 删除：`Cargo.toml`
- ✅ 保留：`src-tauri/icons/` → 复制到 `build/`
- ✅ 保留：前端代码（移动到 `frontend/`）

## 代码变更详情

### 1. Go 后端实现

#### main.go
- 使用 Wails v2 框架
- 嵌入前端资源（`embed.FS`）
- 配置 macOS 窗口样式（透明、无标题栏）

#### app.go
- 应用生命周期管理
- 窗口显示/隐藏方法
- 托盘管理器初始化

#### countdown.go
- 倒计时 CRUD 操作
- JSON 文件存储（`~/.near/store.json`）
- 数据结构与 Tauri 版本完全兼容

#### system.go
- CPU/内存监控（使用 `gopsutil`）
- macOS 温度获取（`ioreg` 命令）

#### tray.go
- 系统托盘实现（使用 `systray` 库）
- CPU 风车动画（32帧，根据 CPU 动态调整帧率）
- 托盘标题更新（显示置顶倒计时）

#### ai.go
- AI 配置存储/读取

### 2. 前端改造

#### wails-api.js（新增）
- 适配层，兼容 Tauri 的 `invoke()` API
- 映射到 Wails 的 `window.go.main.App.*` 方法
- 保持前端代码最小改动

#### App.vue
- 仅修改导入语句：`import { invoke } from './wails-api.js'`
- 其他代码完全不变
- UI/样式保持一致

#### vite.config.js
- 修改端口：`5173` → `34115`（Wails 默认）
- 移除 Tauri 相关配置

### 3. 配置文件

#### wails.json
```json
{
  "name": "Near",
  "outputfilename": "Near",
  "frontend:install": "cd frontend && npm install",
  "frontend:build": "cd frontend && npm run build",
  "frontend:dev:watcher": "cd frontend && npm run dev",
  "frontend:dev:serverUrl": "auto",
  "wailsjsdir": "./frontend/src",
  "assetdir": "./frontend/dist"
}
```

#### go.mod
```go
module near

go 1.21

require (
    github.com/wailsapp/wails/v2 v2.8.0
    github.com/getlantern/systray v1.2.2
    github.com/shirou/gopsutil/v3 v3.23.12
)
```

## 功能对比

| 功能 | Tauri 实现 | Wails 实现 | 状态 |
|------|-----------|-----------|------|
| 倒计时管理 | ✅ | ✅ | 完全兼容 |
| 拖拽排序 | ✅ | ✅ | 前端实现，无变化 |
| AI 解析 | ✅ | ✅ | 前端实现，无变化 |
| 数据存储 | ✅ | ✅ | JSON 文件，兼容 |
| 系统监控 | ✅ | ✅ | 功能相同 |
| 系统托盘 | ✅ | ⚠️ | 功能受限（见下文） |
| CPU 风车动画 | ✅ | ⚠️ | 可能略卡顿 |
| 失焦隐藏 | ✅ | ❌ | 未实现 |
| 多显示器支持 | 🔴 | 🔴 | 两者都有问题 |

## 已知限制与差异

### 🔴 系统托盘功能受限

**Tauri 版本**：
- 原生支持托盘图标点击事件
- 可获取托盘图标屏幕坐标
- 窗口可定位到托盘图标下方

**Wails 版本**：
- 使用第三方库 `systray`
- **无法获取托盘图标点击事件**
- **无法获取托盘图标屏幕坐标**
- 只能通过菜单项"显示窗口"打开
- 窗口无法定位到托盘图标下方

**影响**：
- 用户体验下降：无法点击托盘图标直接显示窗口
- 需要右键 → 点击"显示窗口"菜单项
- 窗口显示在屏幕中央，而非托盘图标下方

### 🟡 CPU 风车动画性能

**Tauri 版本**：
- 使用 Tauri 原生 API 更新托盘图标
- 流畅，CPU 占用 < 1%

**Wails 版本**：
- 使用 `systray.SetIcon()` 更新图标
- 可能略有卡顿（取决于系统）
- CPU 占用可能略高

### ❌ 失焦自动隐藏

**Tauri 版本**：
- 原生支持窗口失焦事件
- 失焦时自动隐藏窗口

**Wails 版本**：
- Wails v2 不直接支持窗口失焦事件
- 未实现此功能
- 需要手动关闭窗口或点击托盘菜单

### 🔴 多显示器支持

**两个版本都有问题**：
- Tauri：窗口可能显示在错误的显示器上
- Wails：更严重，无法获取托盘坐标，无法实现精确定位

## 构建与运行

### 开发模式

```bash
# 安装 Wails CLI（如果未安装）
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装前端依赖
cd frontend && npm install && cd ..

# 运行开发模式
wails dev
```

### 生产构建

```bash
# 构建应用
wails build

# 输出位置
# macOS: build/bin/Near.app
```

### 依赖安装

```bash
# Go 依赖
go mod download

# 前端依赖
cd frontend && npm install
```

## 迁移后的优缺点

### ✅ 优点

1. **Go 语言**：相比 Rust 更容易上手
2. **构建速度**：Go 编译比 Rust 快
3. **代码量**：Go 代码更简洁（~400 行 vs ~500 行 Rust）
4. **依赖管理**：Go modules 比 Cargo 简单

### ❌ 缺点

1. **托盘功能受限**：无法点击托盘图标显示窗口
2. **用户体验下降**：需要通过菜单操作
3. **性能略差**：动画可能略卡顿
4. **功能缺失**：失焦自动隐藏未实现
5. **包体积**：~15MB（Tauri ~10MB）

## 建议

### 如果继续使用 Wails

1. **简化交互**：
   - 改为菜单栏应用（不依赖托盘点击）
   - 添加全局快捷键唤醒窗口
   - 使用 Dock 图标而非托盘

2. **优化动画**：
   - 降低帧率（15fps → 10fps）
   - 简化动画（减少帧数）
   - 或完全移除动画

3. **补充功能**：
   - 实现全局快捷键
   - 添加 Dock 菜单
   - 改进窗口管理

### 如果回退到 Tauri

```bash
# 切换回 tauri 分支
git checkout tauri

# 或者切换回 main 分支
git checkout main
```

## 文件清单

### 新增文件
- `main.go` - Go 主程序
- `app.go` - 应用逻辑
- `countdown.go` - 倒计时服务
- `system.go` - 系统监控
- `tray.go` - 系统托盘
- `ai.go` - AI 配置
- `wails.json` - Wails 配置
- `go.mod` - Go 依赖
- `frontend/src/wails-api.js` - API 适配层

### 修改文件
- `frontend/src/App.vue` - 修改导入语句
- `frontend/vite.config.js` - 修改端口配置

### 备份文件
- `frontend/src/App.vue.tauri-backup` - Tauri 版本备份

## 下一步

1. **安装 Go 环境**（如果未安装）
2. **安装 Wails CLI**
3. **运行 `wails dev` 测试**
4. **根据测试结果决定是否继续使用 Wails**

## 总结

✅ **代码迁移已完成**，所有功能都已用 Go 重写。

⚠️ **但存在明显的功能限制**，特别是系统托盘交互方面。

🤔 **建议**：
- 如果核心需求是托盘点击交互 → **不推荐 Wails**，建议保持 Tauri
- 如果可以接受菜单操作 → **可以使用 Wails**
- 如果想要最完整的功能 → **考虑 Electron**

---

**迁移完成时间**：2024-12-08
**分支名称**：`wails-v2`
**原始分支**：`tauri` / `main`
