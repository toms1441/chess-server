package game

import (
	"encoding/json"
	"fmt"

	"github.com/toms1441/chess-server/internal/board"
	"github.com/toms1441/chess-server/internal/order"
)

type Game struct {
	id   string
	cs   [2]*Client
	turn uint8
	done bool
	b    *board.Board
	cmd  map[string]interface{} // command-specific data
}

func NewGame(cl1, cl2 *Client) (*Game, error) {

	if cl1 == nil || cl2 == nil || cl1.W == nil || cl2.W == nil {
		return nil, ErrClientNil
	}

	if cl1.g != nil || cl2.g != nil {
		return nil, ErrGameIsNotNil
	}

	cl1.num = 1
	cl1.id = fmt.Sprintf("Player %d", cl1.num)

	cl2.num = 2
	cl2.id = fmt.Sprintf("Player %d", cl2.num)

	g := &Game{
		cs:   [2]*Client{cl1, cl2},
		turn: 0,
		cmd:  map[string]interface{}{},
	}

	cl1.g, cl2.g = g, g

	g.b = board.NewBoard()

	g.b.Listen(func(p *board.Piece, src board.Point, dst board.Point, ret bool) {
		if ret {
			if p.T == board.PawnB || p.T == board.PawnF {
				if dst.X == 7 || dst.X == 1 {
					c := g.cs[p.Player-1]
					if c != nil {
						x := order.PromoteModel{
							Src: dst,
						}

						g.Update(c, order.Order{
							ID:        order.Promote,
							Parameter: x,
						})
					}
				}
			}
		}
	})

	//g.SwitchTurn()

	return g, nil
}

// SwitchTurn called after a player ends their turn, to notify the other player.
func (g *Game) SwitchTurn() {
	if g.turn == 1 {
		g.turn = 2
	} else {
		g.turn = 1
	}
	t := g.turn
	// TODO: make this not byte me in the ass
	x, _ := json.Marshal(order.TurnModel{
		Player: t,
	})

	g.UpdateAll(order.Order{ID: order.Turn, Data: x})
}

func (g *Game) IsTurn(c *Client) bool {
	return c.num == g.turn
}

// Update is used to send updates to the client, such as a movement of a piece.
func (g *Game) Update(c *Client, u order.Order) error {
	if u.Data == nil {
		x, ok := ubs[u.ID]
		if !ok {
			return ErrUpdateNil
		}

		err := x(c, &u)
		if err != nil {
			return err
		}
	}

	body, err := json.Marshal(u)
	if err != nil {
		return err
	}

	go func() {
		c.W.Write(body)
	}()

	return nil
}

// UpdateAll sends the update to all of the players.
func (g *Game) UpdateAll(o order.Order) error {
	err := g.Update(g.cs[0], o)
	if err != nil {
		return err
	}

	return g.Update(g.cs[1], o)
}

// Board returns a board copy.
func (g *Game) Board() board.Board {
	return *g.b
}