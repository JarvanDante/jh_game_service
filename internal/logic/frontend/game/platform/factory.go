package platform

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"sync"
)

/*
	现在你的游戏服务使用的是新的接口+工厂模式架构，更容易扩展和维护。如果以后要添加其他游戏平台（如AG、PP等），只需要：

	在platform/目录下创建新的文件夹（如ag/、pp/）
	实现GamePlatform接口
	在init()函数中注册到工厂
	无需修改其他代码
*/
// PlatformFactory 游戏平台工厂
type PlatformFactory struct {
	platforms map[string]func() GamePlatform
	mu        sync.RWMutex
}

var (
	factory *PlatformFactory
	once    sync.Once
)

// GetFactory 获取工厂单例
func GetFactory() *PlatformFactory {
	once.Do(func() {
		factory = &PlatformFactory{
			platforms: make(map[string]func() GamePlatform),
		}
	})
	return factory
}

// Register 注册游戏平台
// platformName: 平台名称，如 "ONE", "AG", "PP"
// creator: 创建平台实例的函数
func (f *PlatformFactory) Register(platformName string, creator func() GamePlatform) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.platforms[platformName] = creator
}

// Create 创建游戏平台实例
func (f *PlatformFactory) Create(platformName string) (GamePlatform, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	creator, exists := f.platforms[platformName]
	ctx := context.Background()
	for k := range f.platforms {
		g.Log().Infof(ctx, "key: %s", k)
	}
	if !exists {
		return nil, fmt.Errorf("不支持的游戏平台: %s", platformName)
	}

	return creator(), nil
}

// GetSupportedPlatforms 获取所有支持的平台列表
func (f *PlatformFactory) GetSupportedPlatforms() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	platforms := make([]string, 0, len(f.platforms))
	for name := range f.platforms {
		platforms = append(platforms, name)
	}
	return platforms
}
