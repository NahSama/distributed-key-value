package raft_handler

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"net/http"
)

type requestJoin struct {
	NodeId      string `json:"node_id"`
	RaftAddress string `json:"raft_address"`
}

func (h *handler) JoinRaftHandler(eCtx echo.Context) error {
	var form = requestJoin{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	var (
		nodeId      = form.NodeId
		raftAddress = form.RaftAddress
	)

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

	future := h.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(raftAddress), 0, 0)
	if err := future.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("failed to add voter: %s", future.Error().Error()),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s at %s joined successfully", nodeId, raftAddress),
		"data":    h.raft.State(),
	})
}
