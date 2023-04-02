package server

import (
	"distributed-db/server/raft_handler"
	"distributed-db/server/store_handler"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

type srv struct {
	listenAddress string
	raft          *raft.Raft
	echo          *echo.Echo
}

func (s srv) Start() error {
	return s.echo.StartServer(&http.Server{
		Addr:         s.listenAddress,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

func New(listenAddress string, badgerDb *badger.DB, r *raft.Raft) *srv {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	// Raft server
	raftHandler := raft_handler.New(r)
	e.POST("/raft/join", raftHandler.JoinRaftHandler)
	e.POST("/raft/remove", raftHandler.RemoveRaftHandler)
	e.GET("/raft/stats", raftHandler.StatsRaftHandler)

	storeHandler := store_handler.New(r, badgerDb)
	e.POST("/store", storeHandler.Store)
	e.GET("/store/:key", storeHandler.Get)
	e.DELETE("/store/:key", storeHandler.Delete)
	e.GET("/store/follower-read/:key", storeHandler.FollowerRead)

	return &srv{
		listenAddress: listenAddress,
		echo:          e,
		raft:          r,
	}
}
