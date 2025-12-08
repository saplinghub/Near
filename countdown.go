package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Countdown 倒计时结构体
type Countdown struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Date      string `json:"date"`
	StartDate string `json:"startDate"`
	IconType  string `json:"iconType"`
	Order     int    `json:"order"`
	Pinned    bool   `json:"pinned"`
	CreatedAt string `json:"createdAt"`
}

// CountdownService 倒计时服务
type CountdownService struct {
	dataPath string
}

// NewCountdownService 创建倒计时服务
func NewCountdownService() *CountdownService {
	homeDir, _ := os.UserHomeDir()
	dataPath := filepath.Join(homeDir, ".near", "store.json")

	// 确保目录存在
	os.MkdirAll(filepath.Dir(dataPath), 0755)

	return &CountdownService{dataPath: dataPath}
}

// loadStore 加载存储数据
func (s *CountdownService) loadStore() (map[string]interface{}, error) {
	data, err := os.ReadFile(s.dataPath)
	if err != nil {
		// 文件不存在，返回空 map
		return make(map[string]interface{}), nil
	}

	var store map[string]interface{}
	if err := json.Unmarshal(data, &store); err != nil {
		return make(map[string]interface{}), err
	}

	return store, nil
}

// saveStore 保存存储数据
func (s *CountdownService) saveStore(store map[string]interface{}) error {
	bytes, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataPath, bytes, 0644)
}

// GetCountdowns 获取所有倒计时
func (a *App) GetCountdowns() ([]Countdown, error) {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return []Countdown{}, err
	}

	countdowns := []Countdown
	if list, ok := store["countdowns"].([]interface{}); ok {
		for _, item := range list {
			bytes, _ := json.Marshal(item)
			var countdown Countdown
			if err := json.Unmarshal(bytes, &countdown); err == nil {
				countdowns = append(countdowns, countdown)
			}
		}
	}

	return countdowns, nil
}

// SaveCountdown 保存倒计时
func (a *App) SaveCountdown(countdown Countdown) error {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return err
	}

	// 获取现有列表
	list := []Countdown{}
	if existing, ok := store["countdowns"].([]interface{}); ok {
		for _, item := range existing {
			bytes, _ := json.Marshal(item)
			var c Countdown
			if err := json.Unmarshal(bytes, &c); err == nil {
				// 如果 ID 相同，跳过（稍后添加新的）
				if c.ID != countdown.ID {
					list = append(list, c)
				}
			}
		}
	}

	// 添加新的或更新的倒计时
	list = append(list, countdown)

	// 保存
	store["countdowns"] = list
	return a.countdownService.saveStore(store)
}

// DeleteCountdown 删除倒计时
func (a *App) DeleteCountdown(id string) error {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return err
	}

	// 过滤掉要删除的项
	list := []Countdown{}
	if existing, ok := store["countdowns"].([]interface{}); ok {
		for _, item := range existing {
			bytes, _ := json.Marshal(item)
			var c Countdown
			if err := json.Unmarshal(bytes, &c); err == nil {
				if c.ID != id {
					list = append(list, c)
				}
			}
		}
	}

	store["countdowns"] = list
	return a.countdownService.saveStore(store)
}

// PinCountdown 置顶倒计时
func (a *App) PinCountdown(id string) ([]Countdown, error) {
	store, err := a.countdownService.loadStore()
	if err != nil {
		return []Countdown{}, err
	}

	// 获取现有列表
	list := []Countdown{}
	if existing, ok := store["countdowns"].([]interface{}); ok {
		for _, item := range existing {
			bytes, _ := json.Marshal(item)
			var c Countdown
			if err := json.Unmarshal(bytes, &c); err == nil {
				// 取消所有置顶
				c.Pinned = false
				// 如果是目标 ID，设置为置顶
				if c.ID == id {
					c.Pinned = true
				}
				list = append(list, c)
			}
		}
	}

	store["countdowns"] = list
	if err := a.countdownService.saveStore(store); err != nil {
		return []Countdown{}, err
	}

	return list, nil
}

// CalculateDays 计算剩余天数
func CalculateDays(targetDate string) int {
	target, err := time.Parse("2006-01-02T15:04", targetDate)
	if err != nil {
		return 0
	}

	now := time.Now()
	// 只比较日期部分
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	targetDateOnly := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, target.Location())

	diff := targetDateOnly.Sub(nowDate)
	days := int(diff.Hours() / 24)

	if days < 0 {
		return 0
	}
	return days
}
