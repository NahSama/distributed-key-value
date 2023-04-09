package raft_handler

import (
	"fmt"
	raft "github.com/NahSama/raft-modified"
	"github.com/labstack/echo/v4"
	"net/http"
)

type requestRemove struct {
	NodeId string `json:"node_id"`
}

func (h *handler) RemoveRaftHandler(eCtx echo.Context) error {
	var form = requestRemove{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	var nodeId = form.NodeId
	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not a leader",
		})
	}

	configFuture := h.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("failed to get raft configuration: %s", err.Error()),
		})
	}

	future := h.raft.RemoveServer(raft.ServerID(nodeId), 0, 0)
	if err := future.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("failed to remove node: %s", future.Error().Error()),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s removed successfully", nodeId),
		"data":    h.raft.State(),
	})
}
