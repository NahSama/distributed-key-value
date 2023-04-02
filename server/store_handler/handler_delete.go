package store_handler

import (
	"distributed-db/fsm"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

func (h handler) Delete(eCtx echo.Context) error {
	var key = strings.TrimSpace(eCtx.Param("key"))

	//
	if key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "DELETE",
		Key:       key,
		Value:     nil,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error removing data in raft cluster: %s\n", err.Error()),
		})
	}

	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error response does not match apply response"),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "successfully remove data",
		"data": map[string]interface{}{
			"key":   key,
			"value": nil,
		},
	})
}
