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
	SiteID        int    `json:"site_id"`
	Code          string `json:"code"`
	Currency      string `json:"currency"`
	DbLinkSite    string `json:"db_link_site"`
	DbLinkBalance string `json:"db_link_balance"`
	DbLinkGame    string `json:"db_link_game"`
	Status        int    `json:"status"`
	Timezone      string `json:"timezone"`
	Version       int64  `json:"version"`
	PublishedAt   string `json:"published_at"`
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

	for _, pair := range pairs {
		// 只处理 sites/{code}/base 格式的key
		if len(pair.Key) > 6 && pair.Key[len(pair.Key)-5:] == "/base" {
			if err := m.updateSiteConfig(pair.Key, pair.Value); err != nil {
				g.Log().Errorf(m.ctx, "更新站点配置失败: key=%s, error=%v", pair.Key, err)
			}
		}
	}

	g.Log().Infof(m.ctx, "Game服务加载了 %d 个站点配置", len(m.siteConfigs))
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

// updateSiteConfig 更新站点配置
func (m *DBConfigManager) updateSiteConfig(key string, value []byte) error {
	var config SiteConfig
	if err := json.Unmarshal(value, &config); err != nil {
		return fmt.Errorf("解析站点配置失败: %v", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查配置是否有变化
	oldConfig, exists := m.siteConfigs[config.Code]
	if exists && oldConfig.Version >= config.Version {
		return nil // 配置没有变化或版本更旧
	}

	// 更新配置
	m.siteConfigs[config.Code] = &config

	// 动态注册数据库配置组
	if err := m.createGameDBConnection(config.Code, &config); err != nil {
		g.Log().Errorf(m.ctx, "注册Game数据库配置组失败: site=%s, error=%v", config.Code, err)
		return err
	}

	g.Log().Infof(m.ctx, "Game服务站点配置已更新: site=%s, version=%d", config.Code, config.Version)
	return nil
}

// createGameDBConnection 为站点创建Game数据库连接
func (m *DBConfigManager) createGameDBConnection(siteCode string, config *SiteConfig) error {
	// 如果没有Game数据库连接配置，跳过
	if config.DbLinkGame == "" {
		g.Log().Warningf(m.ctx, "站点 %s 没有配置Game数据库连接", siteCode)
		return nil
	}

	// 创建数据库配置组名称
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 动态设置数据库配置
	gdb.AddConfigNode(groupName, gdb.ConfigNode{
		Link: config.DbLinkGame,
	})

	g.Log().Infof(m.ctx, "已注册Game数据库配置组: %s -> %s", groupName, config.DbLinkGame)
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
		if int64(config.SiteID) == siteID {
			return code
		}
	}

	return ""
}
