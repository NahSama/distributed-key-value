package main

import (
	"distributed-db/fsm"
	"distributed-db/global"
	raft_badgerdb "distributed-db/raft-badgerdb"
	"distributed-db/server"
	"fmt"
	raft "github.com/NahSama/raft-modified"
	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

// config configuration

const (
	serverPort = "SERVER_PORT"

	raftNodeId = "RAFT_NODE_ID"
	raftPort   = "RAFT_PORT"
	raftVolDir = "RAFT_VOL_DIR"
)

var confKeys = []string{
	serverPort,

	raftNodeId,
	raftPort,
	raftVolDir,
}

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

// main entry point of application start
// run using CONFIG=config.yaml ./program
func main() {

	var v = viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	conf := global.Config{
		Server: global.ConfigServer{
			Port: v.GetInt(serverPort),
		},
		Raft: global.ConfigRaft{
			NodeId:    v.GetString(raftNodeId),
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
	}
	global.GlobalConfig = conf
	log.Printf("%+v\n", conf)

	// Preparing badgerDB
	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	var raftBinAddr = fmt.Sprintf(":%d", conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)

	store, err := raft_badgerdb.NewBadgerStore(filepath.Join(conf.Raft.VolumeDir, "raft.dataRepo"))
	if err != nil {
		log.Fatal(err)
		return
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		log.Fatal(err)
		return
	}

	snapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("raftaddress %s", raftBinAddr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1"+raftBinAddr)
	if err != nil {
		log.Fatal("After tpcAddr")
		return
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		log.Printf("tcpAdress IP %s", tcpAddr.IP)
		log.Printf("After transport")
		log.Fatal(err)
		return
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
		return
	}

	// always start single server as a leader
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(conf.Raft.NodeId),
				Address: transport.LocalAddr(),
			},
		},
	}

	raftServer.BootstrapCluster(configuration)

	srv := server.New(fmt.Sprintf(":%d", conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	return
}
