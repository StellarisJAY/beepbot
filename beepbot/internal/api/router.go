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
func SetupRouter(
	providerService *service.ProviderService,
	agentService *service.AgentService,
	botService *service.BotService,
	sessionService *service.SessionService,
	cronService *service.CronService,
	skillService *service.SkillService,
	mcpService *service.MCPService,
	authService *service.AuthService,
	teamService *service.TeamService,
) *gin.Engine {
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
	mcpHandler := NewMCPHandler(mcpService)
	authHandler := NewAuthHandler(authService)
	teamHandler := NewTeamHandler(teamService)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由（不需要认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		protected := v1.Group("")
		protected.Use(AuthMiddleware(authService))
		{
			// 认证相关
			protected.GET("/auth/me", authHandler.GetMe)
			protected.PUT("/auth/password", authHandler.ChangePassword)
			protected.PUT("/auth/username", authHandler.ChangeUsername)

			// 供应商管理
			providers := protected.Group("/providers")
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
			agents := protected.Group("/agents")
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
				agents.GET("/:id/usage", agentHandler.GetAgentUsageStats)
			}

			// 机器人管理
			bots := protected.Group("/bots")
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
			sessions := protected.Group("/sessions")
			{
				sessions.GET("/:id/messages", sessionHandler.GetSessionMessages)
			}

			// 定时任务管理
			crons := protected.Group("/crons")
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
			skills := protected.Group("/skills")
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

			// MCP 服务器管理
			mcp := protected.Group("/mcp")
			{
				mcp.GET("", mcpHandler.ListMCPServers)
				mcp.GET("/:id", mcpHandler.GetMCPServer)
				mcp.POST("", mcpHandler.CreateMCPServer)
				mcp.PUT("/:id", mcpHandler.UpdateMCPServer)
				mcp.DELETE("/:id", mcpHandler.DeleteMCPServer)
				mcp.PUT("/:id/start", mcpHandler.StartMCPServer)
				mcp.PUT("/:id/stop", mcpHandler.StopMCPServer)
				mcp.POST("/:id/test", mcpHandler.TestMCPConnection)
				mcp.GET("/:id/tools", mcpHandler.GetMCPServerTools)
				mcp.POST("/:id/reconnect", mcpHandler.ReconnectMCPServer)
			}

			// 团队管理
			teams := protected.Group("/teams")
			{
				teams.GET("", teamHandler.ListTeams)
				teams.GET("/:id", teamHandler.GetTeam)
				teams.POST("", teamHandler.CreateTeam)
				teams.PUT("/:id", teamHandler.UpdateTeam)
				teams.DELETE("/:id", teamHandler.DeleteTeam)
				teams.PUT("/:id/status", teamHandler.UpdateTeamStatus)
				teams.GET("/agent/:agent_id", teamHandler.GetAgentTeams)
			}
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
