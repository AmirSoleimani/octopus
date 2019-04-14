package octop

import (
	"bufio"
	"net"

	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

//Peer struct
type Peer struct {
	net.Conn

	Incoming chan []byte
	Outgoing chan []byte
	Disc     chan error

	reader *respreader.RESPReader
	writer *bufio.Writer
}

//NewPeer create new peer
func NewPeer(conn net.Conn) *Peer {

	// reader-writer
	writer := bufio.NewWriter(conn)
	reader := respreader.NewReader(conn)

	// create peer struct
	p := &Peer{
		Conn:     conn,
		Incoming: make(chan []byte),
		Outgoing: make(chan []byte),
		Disc:     make(chan error),
		reader:   reader,
		writer:   writer,
	}

	return p
}

//Start start reader and writer
func (p *Peer) Start() {
	go p.Reader()
	p.Writer()
}

// Close peer
func (p *Peer) Close() error {
	logrus.Errorln("close peer")
	close(p.Incoming)
	close(p.Outgoing)
	p.Conn.Close()

	return nil
}

//Reader reader
func (p *Peer) Reader() {
	for {
		msg, err := p.reader.ReadObject()
		if err != nil {
			p.Disc <- err
			break
		}
		p.Incoming <- msg
	}
}

//Writer writer
func (p *Peer) Writer() {
	for data := range p.Outgoing {
		_, err := p.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		p.writer.Flush()
	}
}
