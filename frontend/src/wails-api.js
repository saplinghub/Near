// Wails API 适配层
// 模拟 Tauri 的 invoke 函数，实际调用 Wails 绑定的 Go 方法

// 注意：这些函数会在 Wails 构建后自动生成
// 这里提供类型定义和占位实现

export async function GetCountdowns() {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.GetCountdowns();
  }
  return [];
}

export async function SaveCountdown(countdown) {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.SaveCountdown(countdown);
  }
}

export async function DeleteCountdown(id) {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.DeleteCountdown(id);
  }
}

export async function PinCountdown(id) {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.PinCountdown(id);
  }
  return [];
}

export async function GetSystemStats() {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.GetSystemStats();
  }
  return {
    cpu: 0,
    memory: 0,
    temperature: 'N/A',
    total_memory: 0,
    used_memory: 0
  };
}

export async function GetAIConfig() {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.GetAIConfig();
  }
  return { baseURL: '', apiKey: '', model: '' };
}

export async function SaveAIConfig(config) {
  if (window.go && window.go.main && window.go.main.App) {
    return await window.go.main.App.SaveAIConfig(config);
  }
}

// 兼容 Tauri 的 invoke 函数
export async function invoke(command, args = {}) {
  const commandMap = {
    'get_countdowns': GetCountdowns,
    'save_countdown': () => SaveCountdown(args.countdown),
    'delete_countdown': () => DeleteCountdown(args.id),
    'pin_countdown': () => PinCountdown(args.id),
    'get_system_stats': GetSystemStats,
    'get_ai_config': GetAIConfig,
    'save_ai_config': () => SaveAIConfig(args.config),
  };

  const fn = commandMap[command];
  if (fn) {
    return await fn();
  }

  throw new Error(`Unknown command: ${command}`);
}
