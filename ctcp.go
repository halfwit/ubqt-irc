package main

import (
	"strings"
	"time"

	"github.com/go-irc/irc"
)

func parseForCTCP(c *irc.Client, m *irc.Message, s *server) {
	prefix := &irc.Prefix{
		Name: c.CurrentNick(),
	}
	token := strings.Split(m.Params[1], " ")
	switch token[0] {
	case "ACTION":
		m.Params[1] = strings.Join(token[1:], " ")
		fn := faction
		if m.Params[0] == prefix.Name {
			fn = fselfaction
		}
		feed(fn, m.Params[0], s, m)
	case "CLIENTINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "CLIENTINFO",
			Params:  []string{m.Prefix.Name, "ACTION CLIENTINFO FINGER PING SOURCE TIME USER INFO VERSION"},
		})
		feed(fserver, "server", s, m)
	case "FINGER":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "FINGER",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		feed(fserver, "server", s, m)
	case "PING", "PING":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "PONG",
			Params:  []string{m.Prefix.Name, m.Params[1]},
		})
	case "SOURCE":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "SOURCE",
			Params:  []string{m.Prefix.Name, "https://github.com/altid/ircfs"},
		})
		feed(fserver, "server", s, m)
	case "TIME":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "TIME",
			Params:  []string{m.Prefix.Name, time.Now().Format(time.RFC3339)},
		})
		feed(fserver, "server", s, m)
	case "VERSION":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "VERSION",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		feed(fserver, "server", s, m)
	case "USERINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "USERINFO",
			Params:  []string{m.Prefix.Name, c.CurrentNick()},
		})
		feed(fserver, "server", s, m)
	default:
		// BUG(halfwit): When we write to a channel from another connected client
		// such as is possible over ZNC, channels aren't created properly
		// We'll have to validate channels are created for any log requests.
		if strings.Contains(m.Params[1], c.CurrentNick()) {
			feed(fhighlight, m.Params[0], s, m)
			s.m <- &msg{
				fn:   fnotification,
				buff: m.Params[0],
				from: m.Prefix.Name,
				data: m.Trailing(),
			}
			return
		}
		// PM, make sure the file exists
		if m.Params[0] == c.CurrentNick() {
			s.j <- m.Prefix.Name
			feed(fbuffer, m.Prefix.Name, s, m)
			return
		}
		if c.FromChannel(m) {
			s.j <- m.Params[0]
			feed(fbuffer, m.Params[0], s, m)
			return
		}
		feed(fserver, "server", s, m)
	}
}