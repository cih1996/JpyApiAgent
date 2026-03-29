package centControlPlatform

import (
	"adminApi"

	"cnb.cool/accbot/goTool/sessionPkg"
	"github.com/ghp3000/logs"
)

// GetApi 返回内部的 AdminApi 实例，供外部适配层访问。
func (c *Core) GetApi() *adminApi.AdminApi {
	return c.server.api
}

// GetSession 返回内部的 Session 实例，供外部适配层访问。
func (c *Core) GetSession() *sessionPkg.Session {
	return c.server.session
}

// Reset 重置 Core 状态（用于强制重新登录）
func (c *Core) Reset() {
	logs.Info("[Core.Reset] 开始重置 Core 状态...")
	if c.server.session != nil {
		logs.Info("[Core.Reset] 关闭旧的 session 连接...")
		c.server.session.Close("forceRelogin")
		c.server.session = nil
		c.server.api = nil
	}
	c.token = ""
	c.Devices.Range(func(key, value any) bool {
		c.Devices.Delete(key)
		return true
	})
	c.MiddleAgents.Range(func(key, value any) bool {
		c.MiddleAgents.Delete(key)
		return true
	})
	c.middleWareId.Range(func(key, value any) bool {
		c.middleWareId.Delete(key)
		return true
	})
	c.PortMaps.Range(func(key, value any) bool {
		c.PortMaps.Delete(key)
		return true
	})
	logs.Info("[Core.Reset] Core 状态已重置")
}
