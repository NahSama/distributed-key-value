package raft_handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h *handler) StatsRaftHandler(eCtx echo.Context) error {
	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "Raft status",
		"data":    h.raft.Stats(),
	})
}
