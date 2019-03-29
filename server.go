package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path"

	"github.com/go-irc/irc"
	"github.com/ubqt-systems/fslib"
)

type server struct {
	conn   net.Conn
	conf   irc.ClientConfig
	cert   tls.Certificate
	e      chan string // events
        j      chan string // joins
	m      chan *msg   // messages
	done   chan struct{}
	addr   string
	buffs  string
	filter string
	log    string
	port   string
	ssl    string
}

func newServer(c *config) *server {
	m := make(chan *msg)
	e := make(chan string)
	j := make(chan string)
	s := &server{
		e:	e,
		m:      m,
		j:	j,
		addr:   c.addr,
		buffs:  c.chans,
		cert:   c.cert,
		filter: c.filter,
		log:    c.log,
		port:   c.port,
		ssl:    c.ssl,
	}
	conf := irc.ClientConfig{
		User:    c.user,
		Nick:    c.nick,
		Name:    c.name,
		Pass:    c.pass,
		Handler: handlerFunc(s),
	}
	s.conf = conf
	return s
}

func (s *server) Open(c *fslib.Control, name string) error {
	err := c.CreateBuffer(name, "feed")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.conn, "JOIN %s\n", name)
	return err
}

func (s *server) Close(c *fslib.Control, name string) error {
	err := c.DeleteBuffer(name, "feed")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.conn, "PART %s\n", name)
	return err
}

func (s *server) Default(c *fslib.Control, cmd, from, msg string) error {
	switch cmd {
	case "a", "act", "action", "me":
		return action(s, from, msg)
	case "msg", "query":
		return pm(s, msg)
	case "nick":
		// Make sure we update s.conf.Name when we update username
		s.conf.Name = msg
		fmt.Fprintf(s.conn, "NICK %s\n", msg)
		return nil
	}
	return fmt.Errorf("Unknown command %s", cmd)
}

// input is always sent down raw to the server
func (s *server) Handle(bufname, message string) error {
	buffer := path.Base(bufname)
	_, err := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, buffer, message)
	s.m <- &msg{
		buff: buffer,
		from: s.conf.Nick,
		data: message,
		fn:   fself,
	}
	return err
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fslib.Control) {
	for {
		select {
		case e := <- s.e:
			c.Event(e)
		case j := <- s.j:
			buffs := getChans(j)
			for _, buff := range buffs {
				if ! c.HasBuffer(buff, "feed") {
					s.Open(c, buff)
				}
			}
		case m := <- s.m:
			fileWriter(c, m)
		case <- ctx.Done():			
			s.conn.Close()
			return
		}
	}

}

func (s *server) connect(ctx context.Context) error {
	var tlsConfig *tls.Config
	dialString := s.addr + ":" + s.port
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", dialString)
	if err != nil {
		return err
	}
	switch s.ssl {
	case "simple":
		tlsConfig = &tls.Config{
			ServerName:         dialString,
			InsecureSkipVerify: true,
		}
	case "certificate":
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				s.cert,
			},
			ServerName:   dialString,
		}

	default:
		s.conn = conn
		return nil
	}
	tlsconn := tls.Client(conn, tlsConfig)
	tlsconn.Handshake()
	s.conn = tlsconn
	return nil
}
