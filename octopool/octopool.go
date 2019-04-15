package octop

import (
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

//OctoPool struct
type OctoPool struct {
	sync.RWMutex
	Queue chan *Peer

	InitCap int
	MaxCap  int

	Factory Factory
}

//Factory type
type Factory func() (net.Conn, error)

//NewOctoPool new octo pool
func NewOctoPool(initCap, maxCap int, factory Factory) (*OctoPool, error) {

	// create octopool
	op := &OctoPool{
		Queue:   make(chan *Peer, maxCap),
		InitCap: initCap,
		MaxCap:  maxCap,
		Factory: factory,
	}

	// initialize peer
	for i := 0; i < initCap; i++ {
		peer, err := op.createPeer()
		if err != nil {
			logrus.Fatalln(err)
		}
		op.Queue <- peer
	}

	return op, nil
}

func (o *OctoPool) createPeer() (*Peer, error) {

	// create connection
	conn, err := o.Factory()
	if err != nil {
		return nil, err
	}

	// create peer > start > add to queue
	peer := NewPeer(conn)
	go peer.Start()
	return peer, nil
}

//Get get
func (o *OctoPool) Get() (*Peer, error) {

	// rw lock
	o.RLock()
	defer o.RUnlock()

	select {
	case p := <-o.Queue:
		if p == nil {
			return nil, ErrClosed
		}

		return p, nil
	default:
		// TODO: watch
		return o.createPeer()
	}
}

//Put put back connection into queue
func (o *OctoPool) Put(p *Peer) error {

	if p == nil {
		return ErrClosed
	}

	// rw lock
	o.Lock()
	defer o.Unlock()

	select {
	case o.Queue <- p:
		return nil
	default:
		logrus.Warnln("pool is full, close connection")
		return p.Close()
	}
}
