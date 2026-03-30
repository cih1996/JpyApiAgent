package middleAgentRtc

import (
	"fmt"
	"time"

	"github.com/ghp3000/logs"
)

// ReconnectForSync 是原始 ApiAgent 的薄扩展，用于同步命令失败后的单次重连。
func (s *TypeInfo) ReconnectForSync(logMsg string) error {
	logs.Warn("[ApiAgent 原始扩展] 中间件[%d]%s首次同步失败，尝试重建 RTC 后重试", s.MiddlewareId, logMsg)

	if s.Rtc != nil {
		_ = s.Rtc.Close()
		s.Rtc = nil
	}
	s.Code = 0

	token, err := s.GetToken(s.MiddlewareId)
	if err != nil {
		return fmt.Errorf("中间件[%d]重取RtcToken失败: %v", s.MiddlewareId, err)
	}
	s.TokenInfo = token
	s.Connect()

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if s.Rtc != nil && s.Code > 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("中间件[%d]重连后仍未就绪(code=%d)", s.MiddlewareId, s.Code)
}

// ShutdownForReset 用于主动清理/重登/切换环境时停用 middle RTC。
func (s *TypeInfo) ShutdownForReset() {
	s.Reconnect = 1
	if s.Rtc != nil {
		_ = s.Rtc.Close()
	}
	s.Rtc = nil
	s.Code = 0
	del(s.MiddlewareId)
	if s.DelMiddleware != nil {
		s.DelMiddleware(s.MiddlewareId)
	}
	s.WsClients.Range(func(key, value any) bool {
		if client, ok := value.(*WsClient); ok {
			_ = client.Close()
		}
		s.WsClients.Delete(key)
		return true
	})
}
