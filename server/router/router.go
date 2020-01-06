// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/dimfeld/httptreemux"
	"github.com/gin-gonic/gin"

	"github.com/coderun-top/coderun/server/router/middleware/ginrus"
	"github.com/coderun-top/coderun/server/router/middleware/header"
	"github.com/coderun-top/coderun/server/router/middleware/session"
	"github.com/coderun-top/coderun/server/router/middleware/token"
	"github.com/coderun-top/coderun/server/server"
	"github.com/coderun-top/coderun/server/server/debug"
	"github.com/coderun-top/coderun/server/server/metrics"
	// "github.com/coderun-top/coderun/server/server/web"
)

// Load loads the router
func Load(mux *httptreemux.ContextMux, middleware ...gin.HandlerFunc) http.Handler {

	e := gin.New()
	e.Use(gin.Recovery())

	e.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true))

	e.Use(header.NoCache)
	e.Use(header.Options)
	e.Use(header.Secure)
	e.Use(middleware...)
	// e.Use(session.SetUser())
	e.Use(token.Refresh)

	// e.NoRoute(func(c *gin.Context) {
	// 	req := c.Request.WithContext(
	// 		web.WithUser(
	// 			c.Request.Context(),
	// 			session.User(c),
	// 		),
	// 	)
	// 	mux.ServeHTTP(c.Writer, req)
	// })

	//e.GET("/logout", server.GetLogout)
	//e.GET("/login", server.HandleLogin)
	e.GET("/api/dig", server.Dig)
	e.POST("/api/test", server.Test)

	project := e.Group("/api/project/:projectname")
	{
		project.Use(token.TokenVerify(server.Config.Services.Grpc))

		project.GET("", server.GetSelf)
		// init project, create token, default charts, default images repository
		project.POST("", server.InitProject)
		project.GET("/info", server.GetProjectInfo)

		project.GET("/agent_key", server.GetAgentKey)
		project.GET("/agents", server.GetAgents)
		// project.PUT("/agent", server.PutAgent)
		// project.PUT("/agent/:status", server.PutAgentStatus)

		// 获取当前Project下所有激活的仓库
		project.GET("/activated/repos", server.GetRepos)

		project.POST("/token/:tokenname", server.PostToken)
		project.DELETE("/token/:id", server.DeleteToken)
		project.PUT("/token/:id", server.PatchToken)
		project.GET("/token", server.GetToken)

		project.GET("registry", server.GetRegistryList)
		project.POST("registry", server.PostRegistry)
		project.GET("/registry/:name", server.GetRegistry)
		project.GET("/inside/registry/:name", server.GetRegistryInside)
		project.PATCH("/registry/:name", server.PatchRegistry)
		project.DELETE("/registry/:name", server.DeleteRegistry)

		project.GET("helm", server.GetHelmList)
		project.POST("helm", server.PostHelm)
		project.GET("/helm/:name", server.GetHelm)
		project.GET("/inside/helm/:name", server.GetHelmInside)
		project.PATCH("/helm/:name", server.PatchHelm)
		project.DELETE("/helm/:name", server.DeleteHelm)

		project.GET("/k8s/cluster", server.GetK8sClusterList)
		project.GET("/k8s/cluster/:name", server.GetK8sCluster)
		project.POST("/k8s/cluster", server.PostK8sCluster)
		project.PUT("/k8s/cluster/:name", server.PutK8sCluster)
		project.DELETE("/k8s/cluster/:name", server.DeleteK8sCluster)

		project.GET("/builds", server.GetProjectBuilds)

		project.GET("/env", server.GetProjectEnv)
		project.PUT("/env", server.PutProjectEnv)
		project.GET("/webhook", server.GetProjectWebHook)
		project.PUT("/webhook", server.PutProjectWebHook)
		project.GET("/email", server.GetProjectEmail)
		project.PUT("/email", server.PutProjectEmail)

		project.GET("/k8s/deploy", server.GetK8sDeployList)
		project.GET("/k8s/deploy/:name", server.GetK8sDeploy)
		project.POST("/k8s/deploy", server.CreateK8sDeploy)
		project.DELETE("/k8s/deploy", server.DeleteK8sDeploy)
	}

	// URL: /api/repo/:projectname/:tokenId?name=coderun-top/startup
	provider := e.Group("/api/repo/:projectname/:id")
	{
		provider.Use(token.TokenVerify(server.Config.Services.Grpc))
		provider.Use(session.SetUser())

		// 根据token名称获取git的仓库列表
		provider.GET("", server.GetRepoName)
		// post add auth
		// repo.POST("", server.PostRepo)
		provider.POST("", session.MustAdmin(server.Config.Services.Grpc), server.PostRepo)

		provider.GET("/branches", server.GetBranches)
		provider.GET("/file", server.EmptyFile)
		provider.GET("/get/file", server.GetFile)
	}

	// URL: /api/repos/:projectname/...?token_name=abc&repo_name=coderun/startup
	repo := e.Group("/api/repos/:projectname")
	{
		repo.Use(token.TokenVerify(server.Config.Services.Grpc))
		repo.Use(session.SetRepoAndUser())

		repo.GET("", server.GetRepo)
		// repo.DELETE("", session.RrojectJurisdiction(1), server.DeleteRepo)
		repo.DELETE("", session.MustAdmin(server.Config.Services.Grpc), server.DeleteRepo)

		repo.POST("/repair", server.RepairRepo)

		repo.GET("/webhook", server.GetPipelineWebHook)
		repo.PUT("/webhook", server.PutPipelineWebHook)

		repo.GET("/email", server.GetPipelineEmail)
		repo.PUT("/email", server.PutPipelineEmail)

		repo.GET("/branches", server.GetRepoBranches)
		repo.GET("/file", server.RepoEmptyFile)
		repo.GET("/get/file", server.GetRepoFile)

		repo.GET("/pipelines", server.GetPipelines)
		// repo.GET("/pipeline/:pipeline_id/env", server.GetPipelineEnv)
		// repo.PUT("/pipeline/:pipeline_id/env", server.PutPipelineEnv)
		repo.GET("/builds", server.GetBuilds)
		repo.POST("/builds", server.CreateBuild)

		// repo.GET("/builds/:number", server.GetBuild)
		// repo.POST("/builds/:number", server.PostBuild)
		// repo.POST("/builds/:number/proc/:pid", server.PostStepBuild)
		// repo.DELETE("/builds/:number", server.ZombieKill)
		// repo.POST("/builds/:number/approve", server.PostApproval)
		// repo.POST("/builds/:number/decline", server.PostDecline)
		// // cancel job
		// repo.DELETE("/builds/:number/job/:job", server.DeleteBuild)
		// // logs
		// repo.GET("/builds/:number/logs/:pid", server.GetProcLogs)
		// repo.DELETE("/builds/:number/logs", server.DeleteBuildLogs)
		// // 暂时不使用
		// // repo.POST("/k8s/deploy/bind", server.K8sDeployBind)
	}

	// URL: /api/repos/:projectname/pipeline/:pipeline_id/env?token_name=abc&repo_name=coderun/startup
	pipeline := e.Group("/api/pipeline/:projectname/:pipeline_id")
	{
		pipeline.Use(token.TokenVerify(server.Config.Services.Grpc))
		pipeline.Use(session.SetRepoAndUser())

		pipeline.PUT("", server.PutPipeline)

		pipeline.GET("/env", server.GetPipelineEnv)
		pipeline.PUT("/env", server.PutPipelineEnv)

		pipeline.GET("/webhook", server.GetPipelineWebHook)
		pipeline.PUT("/webhook", server.PutPipelineWebHook)

		pipeline.GET("/email", server.GetPipelineEmail)
		pipeline.PUT("/email", server.PutPipelineEmail)
	}

	build := e.Group("/api/build/:projectname/:number")
	{
		build.Use(token.TokenVerify(server.Config.Services.Grpc))
		build.Use(session.SetRepoAndUser())

		build.GET("", server.GetBuild)
		build.POST("", server.PostBuild)
		build.DELETE("", server.ZombieKill)

		build.POST("/proc/:pid", server.PostStepBuild)
		build.POST("/approve", server.PostApproval)
		build.POST("/decline", server.PostDecline)
		// cancel job
		build.DELETE("/job/:job", server.DeleteBuild)
		// logs
		build.GET("/logs/:pid", server.GetProcLogs)
		build.DELETE("/logs", server.DeleteBuildLogs)
	}

	// 提供给Agent和Build过程中容器访问的API
	agent := e.Group("/api/agent/:projectname")
	{
		agent.Use(token.TokenVerify(server.Config.Services.Grpc))
		agent.GET("/registry/:name", server.GetRegistryInside)
		agent.GET("/helm/:name", server.GetHelmInside)
		agent.GET("/k8s/cluster/:name", server.GetK8sClusterInside)
		agent.GET("/k8s/deploy/:name", server.GetK8sDeploy)
	}

	star := e.Group("/api/star")
	{
		star.Use(token.TokenVerify(server.Config.Services.Grpc))
		star.GET("", server.GetStarList2)
		star.POST("", server.PostStar)
		star.DELETE("", server.DeleteStar)
	}

	badges := e.Group("/api/badges/:projectname")
	{
		badges.Use(session.SetRepoAndUser())
		badges.GET("/status.svg", server.GetBadge)
		badges.GET("/cc.xml", server.GetCC)
	}

	helm := e.Group("/api/helm/:username")
	{
		helm.Use(token.TokenVerify(server.Config.Services.Grpc))
		helm.Use(token.GetUserAccessToken(server.Config.Services.Grpc))

		helm.GET("/:prefix", server.GetList)
		//helm.POST("", server.CreateK8sDeploy)

		helm.GET("/:prefix/:name", server.GetHelmName)
		helm.GET("/:prefix/:name/:version", server.GetHelmNameVersion)
		helm.DELETE("/:prefix/:name/:version", server.DeleteHelmNameVersion)

	}

	e.POST("/hook/:provider", server.PostHook)
	e.POST("/api/hook/:provider", server.PostHook)

	sse := e.Group("/stream")
	{
		sse.Use(token.TokenVerifyByQuery(server.Config.Services.Grpc))
		sse.GET("/events/:projectname", server.EventStreamSSE)
		// sse.GET("/events", server.EventStreamSSE)
		sse.GET("/logs/:projectname/:number/:pid",
			session.SetRepoAndUser(),
			//session.SetRepo(),
			//session.SetPerm(),
			//session.MustPull,
			server.LogStreamSSE,
		)
	}

	info := e.Group("/api/info")
	{
		info.GET("/queue",
			//		session.MustAdmin(),
			server.GetQueueInfo,
		)
	}

	//auth := e.Group("/authorize")
	//{
	//	auth.GET("", server.HandleAuth)
	//	auth.POST("", server.HandleAuth)
	//	auth.POST("/token", server.GetLoginToken)
	//}

	builds := e.Group("/api/builds")
	{
		// builds.Use(session.MustAdmin())
		builds.GET("", server.GetBuildQueue)
	}

	debugger := e.Group("/api/debug")
	{
		// debugger.Use(session.MustAdmin())
		debugger.GET("/pprof/", debug.IndexHandler())
		debugger.GET("/pprof/heap", debug.HeapHandler())
		debugger.GET("/pprof/goroutine", debug.GoroutineHandler())
		debugger.GET("/pprof/block", debug.BlockHandler())
		debugger.GET("/pprof/threadcreate", debug.ThreadCreateHandler())
		debugger.GET("/pprof/cmdline", debug.CmdlineHandler())
		debugger.GET("/pprof/profile", debug.ProfileHandler())
		debugger.GET("/pprof/symbol", debug.SymbolHandler())
		debugger.POST("/pprof/symbol", debug.SymbolHandler())
		debugger.GET("/pprof/trace", debug.TraceHandler())
	}

	monitor := e.Group("/metrics")
	{
		monitor.GET("", metrics.PromHandler())
	}

	e.GET("/version", server.Version)
	e.GET("/healthz", server.Health)

	return e
}
