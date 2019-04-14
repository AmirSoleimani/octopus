package network

import (
	"errors"
	"math/rand"
	"net"

	"github.com/quizofkings/octopus/config"
	octop "github.com/quizofkings/octopus/octopool"
	"github.com/sirupsen/logrus"
)

const (
	maxMoved = 5
)

//NetCommands network interface
type NetCommands interface {
	Write(index int, msg []byte) ([]byte, error)
}

//ClusterPool struct
type ClusterPool struct {
	conns  map[string]*octop.OctoPool // addr => pool
	reconn chan net.Conn
}

//New create network ^-^
func New() NetCommands {

	// check nodes count
	if len(config.Reader.Clusters) == 0 {
		logrus.Fatalln("cluster info is empty!")
	}

	// logger
	logrus.Infoln("create node(s) connection")

	var clusterPoolMap = ClusterPool{
		conns:  map[string]*octop.OctoPool{},
		reconn: make(chan net.Conn),
	}

	// do
	for _, cluster := range config.Reader.Clusters {
		for _, node := range cluster.Nodes {
			clusterPoolMap.AddNode(node)
		}
	}

	return &clusterPoolMap
}

//AddNode add new node when redis ASK/MOVED/initialize
func (c *ClusterPool) AddNode(node string) error {

	// check exist
	if _, exist := c.conns[node]; exist {
		return nil
	}

	p, err := octop.NewOctoPool(config.Reader.Pool.InitCap, config.Reader.Pool.MaxCap, func() (net.Conn, error) {
		logrus.Infof("create new connection, remoteAddr:%s", node)
		return net.Dial("tcp", node)
	})
	if err != nil {
		logrus.Fatalln(err)
		return err
	}
	logrus.Infoln(node, "created")
	c.conns[node] = p

	return nil
}

//Write write message into connection
func (c *ClusterPool) Write(clusterIndex int, msg []byte) ([]byte, error) {

	// get cluster node connection from pool
	addr, err := c.getRandomNode(clusterIndex)
	if err != nil {
		return nil, err
	}

	return c.writeAction(addr, msg)
}

func (c *ClusterPool) writeAction(nodeAddr string, msg []byte) ([]byte, error) {

	// variable
	bufNode := []byte{}

	// get node from octopool ^-^
	octoPool := c.conns[nodeAddr]
	peer, err := octoPool.Get()
	if err != nil {
		return nil, err
	}
	defer octoPool.Put(peer)

	// write to peer
	peer.Outgoing <- msg
	for {
		select {
		case bufNode = <-peer.Incoming:
			break
		case err := <-peer.Disc:
			peer.Close()
			return nil, err
		}
		break
	}

	// check moved or ask
	moved, ask, addr := redisHasMovedError(bufNode)
	if moved || ask {
		if err := c.AddNode(addr); err != nil {
			logrus.Errorln(err)
			return nil, err
		}

		return c.writeAction(addr, msg)
	}

	return bufNode, nil
}

func (c *ClusterPool) getRandomNode(clusterIndex int) (string, error) {

	// check requested index
	if clusterIndex > len(config.Reader.Clusters)-1 {
		return "", errors.New("cluster index bigger than registered clusters")
	}

	// get from config
	clusterNodes := config.Reader.Clusters[clusterIndex].Nodes
	lnClusterNodes := len(clusterNodes)
	var choosedNode string
	if lnClusterNodes > 1 {
		// choose randomly
		choosedNode = clusterNodes[rand.Intn(lnClusterNodes)]
	} else {
		choosedNode = clusterNodes[0]
	}

	return choosedNode, nil
}
