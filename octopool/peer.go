package octop

import (
	"bufio"
	"net"

	"github.com/quizofkings/octopus/respreader"
	"github.com/quizofkings/octopus/utils"
	"github.com/sirupsen/logrus"
)

//Peer struct
type Peer struct {
	net.Conn

	Tag      string
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
		Tag:      utils.RandStringRunes(15),
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

	defer func() {
		p.Disc <- ErrClosed
	}()

	go p.Reader()
	p.Writer()
}

// Close peer
func (p *Peer) Close() error {

	// close channels
	close(p.Incoming)
	close(p.Outgoing)

	// close net.Conn
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
	// for {
	// 	message := make([]byte, 4096)
	// 	length, err := p.Conn.Read(message)
	// 	if err != nil {
	// 		p.Disc <- err
	// 		break
	// 	}
	// 	if length > 0 {
	// 		p.Incoming <- message[:length]
	// 	}
	// }
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
