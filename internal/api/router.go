package api

import (
	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置 API 路由
func SetupRouter(providerService *service.ProviderService, agentService *service.AgentService) *gin.Engine {
	router := gin.Default()

	// 创建处理器
	providerHandler := NewProviderHandler(providerService)
	agentHandler := NewAgentHandler(agentService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 供应商管理
		providers := v1.Group("/providers")
		{
			providers.GET("", providerHandler.ListProviders)
			providers.GET("/:id", providerHandler.GetProvider)
			providers.POST("", providerHandler.CreateProvider)
			providers.PUT("/:id", providerHandler.UpdateProvider)
			providers.DELETE("/:id", providerHandler.DeleteProvider)
			providers.PUT("/:id/default", providerHandler.SetDefaultProvider)
			providers.GET("/type/:type", providerHandler.GetProvidersByType)
		}

		// 智能体管理
		agents := v1.Group("/agents")
		{
			agents.GET("", agentHandler.ListAgents)
			agents.GET("/active", agentHandler.GetActiveAgents)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.POST("", agentHandler.CreateAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
			agents.PUT("/:id/status", agentHandler.UpdateAgentStatus)

			// 渠道绑定管理
			agents.GET("/:id/channels", agentHandler.GetChannels)
			agents.POST("/:id/channels", agentHandler.CreateChannel)
			agents.PUT("/:id/channels/:channelId", agentHandler.UpdateChannel)
			agents.DELETE("/:id/channels/:channelId", agentHandler.DeleteChannel)
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
