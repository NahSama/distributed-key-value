package store_handler

import (
	"distributed-db/global"
	"encoding/json"
	"fmt"
	raft "github.com/NahSama/raft-modified"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

func (h *handler) Get(eCtx echo.Context) error {
	key := strings.TrimSpace(eCtx.Param("key"))

	if key == "" {
		return eCtx.JSON(http.StatusUnprocessableEntity, &Response{
			Data: KeyValue{
				Key:   "",
				Value: "",
			},
			Error: errors.New("key is empty"),
		})
	}

	configure := h.raft.GetConfiguration()
	if err := configure.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, &Response{
			Data: KeyValue{
				Key:   "",
				Value: "",
			},
			Error: errors.New(fmt.Sprintf("error when getting raft configuration %s", err.Error())),
		})
	}

	// Current node is leader

	data, err := h.localRead(key)
	if err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, &Response{
			Data: KeyValue{
				Key:   "",
				Value: "",
			},
			Error: errors.New(fmt.Sprintf("Cannot get data %s", err.Error())),
		})
	}

	if h.raft.State() == raft.Leader {
		return eCtx.JSON(http.StatusOK, &Response{
			Data: KeyValue{
				Key:   key,
				Value: data,
			},
			Error: nil,
		})
	}

	// follower read;
	servers := configure.Configuration().Servers
	rand.Shuffle(len(servers), func(i, j int) {
		servers[i], servers[j] = servers[j], servers[i]
	})

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	cur := &KeyValueWithLastIndex{
		Key:       key,
		Value:     data,
		LastIndex: h.raft.LastIndex(),
	}

	cnt := 1
	_, leaderId := h.raft.LeaderWithID()
	for i := 0; i < len(servers); i++ {
		if cnt > len(servers)/2 {
			break
		}
		server := servers[i]
		if server.ID == leaderId {
			continue
		}

		if server.ID == raft.ServerID(global.GlobalConfig.Raft.NodeId) {
			continue
		}
		go func(rServer raft.Server, key string) {
			wg.Add(1)
			fmt.Printf("FollowerRead - Sending request to %s ", rServer.HTTPAddress)
			defer wg.Done()
			url := fmt.Sprintf("http://%s/store/follower-read/%s", rServer.HTTPAddress, key)
			httpResp, err := http.Get(url)
			if err != nil {
				return
			}
			defer httpResp.Body.Close()
			if httpResp.StatusCode != http.StatusOK {
				return
			}

			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return
			}

			var data interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return
			}

			resp, ok := data.(FollowerReadResponse)
			if !ok {
				return
			}

			if err = resp.Error; err != nil {
				return
			}

			// do comparison
			if key != resp.Data.Key {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if resp.Data.LastIndex >= cur.LastIndex {
				cur.Value = resp.Data.Value
				cur.LastIndex = resp.Data.LastIndex
			}

			return
		}(server, key)
		cnt++
	}

	wg.Wait()

	return eCtx.JSON(http.StatusOK, &Response{
		Data: KeyValue{
			Key:   key,
			Value: cur.Value,
		},
		Error: nil,
	})
}

func (h *handler) localRead(key string) (interface{}, error) {
	var keyByte = []byte(key)

	txn := h.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)

	var value = make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	var data interface{}
	if value != nil && len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	return data, err
}
