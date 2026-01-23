package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hashicorp/consul/api"
)

// SiteConfig 站点配置结构
type SiteConfig struct {
	Meta   MetaConfig   `json:"Meta"`
	Base   BaseConfig   `json:"Base"`
	Domain DomainConfig `json:"Domain"`
	Game   []GameConfig `json:"Game"`
}

// MetaConfig 元数据配置
type MetaConfig struct {
	Version     int64  `json:"version"`
	PublishedAt string `json:"published_at"`
	AdminId     int    `json:"admin_id"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	SiteID        int    `json:"site_id"`
	Code          string `json:"code"`
	TimeZone      string `json:"timezone"`
	Currency      string `json:"currency"`
	DbLinkSite    string `json:"db_link_site"`
	DbLinkBalance string `json:"db_link_balance"`
	DbLinkGame    string `json:"db_link_game"`
	Status        int    `json:"status"`
}

// DomainConfig 域名配置
type DomainConfig struct {
	API      []string `json:"api"`
	H5       []string `json:"h5"`
	Admin    []string `json:"admin"`
	Callback []string `json:"callback"`
}

// GameConfig 游戏配置
type GameConfig struct {
	GameId    int    `json:"game_id"`
	Type      int    `json:"type"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	ImageCode string `json:"image_code"`
	DbName    string `json:"db_name"`
	ApiId     string `json:"api_id"`
	ApiKey    string `json:"api_key"`
	ApiUrl    string `json:"api_url"`
}

// DBConfigManager 数据库配置管理器
type DBConfigManager struct {
	consul      *api.Client
	siteConfigs map[string]*SiteConfig // key: site_code, value: config
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	dbManager *DBConfigManager
	once      sync.Once
)

// GetDBConfigManager 获取数据库配置管理器单例
func GetDBConfigManager() *DBConfigManager {
	once.Do(func() {
		dbManager = &DBConfigManager{
			siteConfigs: make(map[string]*SiteConfig),
		}
	})
	return dbManager
}

// Init 初始化数据库配置管理器
func (m *DBConfigManager) Init(consulClient *api.Client) error {
	m.consul = consulClient
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// 初始加载所有站点配置
	if err := m.loadAllSiteConfigs(); err != nil {
		g.Log().Errorf(m.ctx, "初始加载站点配置失败: %v", err)
		return err
	}

	// 启动配置监听
	go m.watchSiteConfigs()

	g.Log().Info(m.ctx, "Game服务数据库配置管理器初始化完成")
	return nil
}

// Stop 停止配置管理器
func (m *DBConfigManager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	g.Log().Info(m.ctx, "Game服务数据库配置管理器已停止")
}

// loadAllSiteConfigs 加载所有站点配置
func (m *DBConfigManager) loadAllSiteConfigs() error {
	kv := m.consul.KV()

	// 获取所有 sites/ 前缀的配置
	pairs, _, err := kv.List("sites/", nil)
	if err != nil {
		return fmt.Errorf("获取站点配置列表失败: %v", err)
	}

	updatedCount := 0
	for _, pair := range pairs {
		updated, err := m.updateSiteConfig(pair.Key, pair.Value)
		if err != nil {
			g.Log().Errorf(m.ctx, "更新站点配置失败: key=%s, error=%v", pair.Key, err)
		} else if updated {
			updatedCount++
		}
	}

	// 删除定时任务的日志输出
	// if updatedCount > 0 {
	// 	g.Log().Infof(m.ctx, "Game服务更新了 %d 个站点配置，当前共 %d 个站点", updatedCount, len(m.siteConfigs))
	// }
	return nil
}

// watchSiteConfigs 监听站点配置变化
func (m *DBConfigManager) watchSiteConfigs() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// 每30秒检查一次配置变化
			time.Sleep(30 * time.Second)
			if err := m.loadAllSiteConfigs(); err != nil {
				g.Log().Errorf(m.ctx, "重新加载站点配置失败: %v", err)
			}
		}
	}
}

// updateSiteConfig 更新站点配置，返回是否有更新
func (m *DBConfigManager) updateSiteConfig(key string, value []byte) (bool, error) {
	var config SiteConfig
	if err := json.Unmarshal(value, &config); err != nil {
		return false, fmt.Errorf("解析站点配置失败: %v", err)
	}

	siteCode := config.Base.Code
	if siteCode == "" {
		return false, fmt.Errorf("站点代码为空")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查配置是否有变化
	oldConfig, exists := m.siteConfigs[siteCode]
	if exists && oldConfig.Meta.Version >= config.Meta.Version {
		return false, nil // 配置没有变化或版本更旧
	}

	// 检查数据库连接是否变化
	dbChanged := false
	if !exists {
		dbChanged = true
	} else {
		if oldConfig.Base.DbLinkGame != config.Base.DbLinkGame {
			dbChanged = true
			g.Log().Infof(m.ctx, "站点 %s Game数据库连接配置发生变化，将重新注册连接池", siteCode)
		}
	}

	// 更新配置到内存
	m.siteConfigs[siteCode] = &config

	// 只在数据库连接变化时才重新注册连接池
	if dbChanged {
		if err := m.createGameDBConnection(siteCode, &config); err != nil {
			g.Log().Errorf(m.ctx, "注册Game数据库配置组失败: site=%s, error=%v", siteCode, err)
			return false, err
		}
		// 删除定时任务的日志输出
		// g.Log().Infof(m.ctx, "Game服务站点配置已更新（含DB连接池）: site=%s, version=%d", siteCode, config.Meta.Version)
	}
	// 删除定时任务的日志输出
	// else {
	// 	g.Log().Debugf(m.ctx, "Game服务站点配置已更新（仅业务配置）: site=%s, version=%d", siteCode, config.Meta.Version)
	// }

	// 同步游戏配置到site_game表（在数据库连接池注册之后）
	if err := m.syncGameConfig(siteCode, &config); err != nil {
		g.Log().Errorf(m.ctx, "同步游戏配置失败: site=%s, error=%v", siteCode, err)
		// 不返回错误，游戏配置同步失败不影响整体流程
	}

	return true, nil
}

// createGameDBConnection 为站点创建Game数据库连接
func (m *DBConfigManager) createGameDBConnection(siteCode string, config *SiteConfig) error {
	// 如果没有Game数据库连接配置，跳过
	if config.Base.DbLinkGame == "" {
		g.Log().Warningf(m.ctx, "站点 %s 没有配置Game数据库连接", siteCode)
		return nil
	}

	// 创建数据库配置组名称
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 解析连接字符串，移除 "mysql:" 前缀
	link := config.Base.DbLinkGame
	if len(link) > 6 && link[:6] == "mysql:" {
		link = link[6:]
	}

	// 动态设置数据库配置
	gdb.AddConfigNode(groupName, gdb.ConfigNode{
		Type: "mysql",
		Link: link,
	})

	// 验证配置并测试连接
	registeredConfig := gdb.GetConfig(groupName)
	if registeredConfig != nil && len(registeredConfig) > 0 {
		testDB := g.DB(groupName)
		if testDB != nil {
			_, err := testDB.GetValue(m.ctx, "SELECT 1")
			if err != nil {
				g.Log().Errorf(m.ctx, "数据库连接测试失败: %s, error: %v", groupName, err)
				return fmt.Errorf("数据库连接测试失败: %v", err)
			}
			// 删除定时任务的日志输出
			// g.Log().Debugf(m.ctx, "数据库连接池注册成功: %s", groupName)
		}
	} else {
		g.Log().Errorf(m.ctx, "配置注册失败: %s", groupName)
		return fmt.Errorf("配置注册失败")
	}

	return nil
}

// GetSiteConfig 获取站点配置
func (m *DBConfigManager) GetSiteConfig(siteCode string) (*SiteConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.siteConfigs[siteCode]
	if !exists {
		return nil, fmt.Errorf("站点配置不存在: site=%s", siteCode)
	}

	return config, nil
}

// GetAllSiteCodes 获取所有站点代码
func (m *DBConfigManager) GetAllSiteCodes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	codes := make([]string, 0, len(m.siteConfigs))
	for code := range m.siteConfigs {
		codes = append(codes, code)
	}

	return codes
}

// GetSiteCodeBySiteID 根据站点ID获取站点代码
func (m *DBConfigManager) GetSiteCodeBySiteID(siteID int64) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for code, config := range m.siteConfigs {
		if int64(config.Base.SiteID) == siteID {
			return code
		}
	}

	return ""
}

// syncGameConfig 同步游戏配置到site_game表
func (m *DBConfigManager) syncGameConfig(siteCode string, config *SiteConfig) error {
	// 获取站点专用数据库连接
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 检查数据库配置是否存在
	if gdb.GetConfig(groupName) == nil {
		// 删除定时任务的日志输出
		// g.Log().Debugf(m.ctx, "站点数据库配置不存在，跳过游戏配置同步: %s", groupName)
		return nil
	}

	db := g.DB(groupName)
	if db == nil {
		g.Log().Warningf(m.ctx, "站点数据库连接不存在，跳过游戏配置同步: %s", groupName)
		return nil
	}

	// 使用事务处理同步
	return db.Transaction(m.ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 删除该站点的所有游戏配置
		_, err := tx.Model("site_game").Where("site_id", config.Base.SiteID).Delete()
		if err != nil {
			return fmt.Errorf("删除旧游戏配置失败: %v", err)
		}

		// 2. 如果没有游戏配置，直接返回
		if len(config.Game) == 0 {
			// 删除定时任务的日志输出
			// g.Log().Infof(ctx, "站点 %s 没有游戏配置，已清空site_game表", siteCode)
			return nil
		}

		// 3. 批量插入新的游戏配置
		insertData := make([]g.Map, 0, len(config.Game))
		for _, game := range config.Game {
			insertData = append(insertData, g.Map{
				"site_id":    config.Base.SiteID,
				"game_id":    game.GameId,
				"type":       game.Type,
				"name":       game.Name,
				"platform":   game.Platform,
				"image_code": game.ImageCode,
				"db_name":    game.DbName,
				"api_id":     game.ApiId,
				"api_key":    game.ApiKey,
				"api_url":    game.ApiUrl,
				"status":     1, // 默认启用
				"sort":       0,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			})
		}

		_, err = tx.Model("site_game").Data(insertData).Insert()
		if err != nil {
			return fmt.Errorf("插入游戏配置失败: %v", err)
		}

		// 删除定时任务的日志输出
		// g.Log().Infof(ctx, "站点 %s 同步了 %d 个游戏配置到site_game表", siteCode, len(config.Game))
		return nil
	})
}
