package store_handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type KeyValueWithLastIndex struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	LastIndex uint64
}

type FollowerReadResponse struct {
	Data  KeyValueWithLastIndex `json:"data"`
	Error error                 `json:"error"`
}

func (h *handler) FollowerRead(eCtx echo.Context) error {
	var key = strings.TrimSpace(eCtx.Param("key"))
	if key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, &FollowerReadResponse{
			Data: KeyValueWithLastIndex{
				Key:       "",
				Value:     "",
				LastIndex: 0,
			},
			Error: errors.New("key is empty"),
		})
	}

	data, err := h.localRead(key)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, &FollowerReadResponse{
			Data: KeyValueWithLastIndex{
				Key:       "",
				Value:     "",
				LastIndex: 0,
			},
			Error: errors.New(fmt.Sprintf("error getting key %s from storage: %s", key, err.Error())),
		})
	}

	return eCtx.JSON(http.StatusOK, &FollowerReadResponse{
		Data: KeyValueWithLastIndex{
			Key:       key,
			Value:     data,
			LastIndex: h.raft.LastIndex(),
		},
		Error: nil,
	})
}
