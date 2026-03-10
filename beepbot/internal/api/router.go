package api

import (
	"net/http"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/gin-gonic/gin"
)

// CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-Requested-With, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "12 hours")

		// 预检请求直接返回
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SetupRouter 设置 API 路由
func SetupRouter(providerService *service.ProviderService, agentService *service.AgentService, botService *service.BotService, sessionService *service.SessionService, cronService *service.CronService, skillService *service.SkillService) *gin.Engine {
	router := gin.Default()

	// 添加 CORS 中间件
	router.Use(corsMiddleware())

	// 创建处理器
	providerHandler := NewProviderHandler(providerService)
	agentHandler := NewAgentHandler(agentService)
	botHandler := NewBotHandler(botService)
	sessionHandler := NewSessionHandler(sessionService)
	cronHandler := NewCronHandler(cronService)
	skillHandler := NewSkillHandler(skillService)

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
			agents.GET("/defaults", agentHandler.GetAgentDefaults)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.POST("", agentHandler.CreateAgent)
			agents.POST("/:id/validate", agentHandler.ValidateAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
			agents.PUT("/:id/status", agentHandler.UpdateAgentStatus)
			agents.GET("/:id/sessions", sessionHandler.GetAgentSessions)
			agents.GET("/:id/skills", agentHandler.GetAgentSkills)
			agents.PUT("/:id/skills", agentHandler.UpdateAgentSkills)
		}

		// 机器人管理
		bots := v1.Group("/bots")
		{
			bots.GET("", botHandler.ListBots)
			bots.GET("/unbound", botHandler.GetUnboundBots)
			bots.GET("/platform/:platform", botHandler.GetBotsByPlatform)
			bots.GET("/:id", botHandler.GetBot)
			bots.POST("", botHandler.CreateBot)
			bots.PUT("/:id", botHandler.UpdateBot)
			bots.DELETE("/:id", botHandler.DeleteBot)
			bots.PUT("/:id/status", botHandler.UpdateBotStatus)
			bots.PUT("/:id/agent", botHandler.BindAgent)
		}

		// 会话管理
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/:id/messages", sessionHandler.GetSessionMessages)
		}

		// 定时任务管理
		crons := v1.Group("/crons")
		{
			crons.GET("", cronHandler.ListCronJobs)
			crons.GET("/:id", cronHandler.GetCronJob)
			crons.POST("", cronHandler.CreateCronJob)
			crons.PUT("/:id", cronHandler.UpdateCronJob)
			crons.DELETE("/:id", cronHandler.DeleteCronJob)
			crons.PUT("/:id/status", cronHandler.UpdateCronJobStatus)
			crons.GET("/agent/:agent_id", cronHandler.GetCronJobsByAgent)
		}

		// 技能管理
		skills := v1.Group("/skills")
		{
			skills.GET("", skillHandler.ListSkills)
			skills.GET("/:id", skillHandler.GetSkill)
			skills.GET("/:id/detail", skillHandler.GetSkillWithFiles)
			skills.GET("/:id/files", skillHandler.GetSkillFiles)
			skills.GET("/:id/files/:fileId", skillHandler.GetSkillFileContent)
			skills.POST("/upload", skillHandler.UploadSkill)
			skills.DELETE("/:id", skillHandler.DeleteSkill)
			skills.PUT("/:id/status", skillHandler.UpdateSkillStatus)
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
