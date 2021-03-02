package game

import (
	"io"

	"github.com/toms1441/chess-server/internal/order"
)

// Client is a struct used for the server to communicate to the client.
type Client struct {
	// W where to write updates
	W   io.WriteCloser
	num uint8 // player 1 or 2??
	id  string
	// G the underlying game
	g *Game
}

func (c *Client) Do(cmd order.Order) error {
	if c.g == nil {
		return ErrGameNil
	}

	x, ok := cbs[cmd.ID]
	if !ok {
		return ErrCommandNil
	}

	return x(c, cmd)
}

func (c *Client) Game() *Game {
	return c.g
}

func (c *Client) LeaveGame() {
	// TODO: c.Do(CmdForfeit)
	c.g = nil
}
