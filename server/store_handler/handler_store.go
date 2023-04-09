package store_handler

import (
	"distributed-db/fsm"
	"encoding/json"
	"fmt"
	raft "github.com/NahSama/raft-modified"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type requestStore struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (h handler) Store(eCtx echo.Context) error {
	var form = requestStore{}

	// cannot bind Post request body to variable
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	form.Key = strings.TrimSpace(form.Key)

	// key is empty
	if form.Key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	// not the leader
	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       form.Key,
		Value:     form.Value,
	}

	data, err := json.Marshal(payload)

	// Cannot marshal the CommandPayload
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error preparing saving data payload: %s", err.Error()),
		})
	}

	// Apply Raft CommandPayload
	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error persisting data in raft cluster: %s", err.Error()),
		})
	}

	//
	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error response is not match apply response"),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success persisting data",
		"data":    form,
	})

}
