package game

import (
	"encoding/json"
	"fmt"

	"github.com/toms1441/chess-server/internal/board"
	"github.com/toms1441/chess-server/internal/model"
)

// Command is a communication structure sent from the client to the server.
// Data needs to be encoded in JSON, and each command has it's own parameters. Defined in order/model.go

type CommandCallback func(c *Client, o model.Order) error

var cbs map[uint8]CommandCallback

func init() {
	cbs = map[uint8]CommandCallback{
		model.OrMove: func(c *Client, o model.Order) error {
			g := c.g
			if !g.IsTurn(c) {
				return ErrIllegalTurn
			}

			if c.inPromotion() {
				return ErrInPromotion
			}

			s := &model.MoveOrder{}

			err := json.Unmarshal(o.Data, s)
			// unmarshal the order
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}

			if !board.BelongsTo(s.ID, c.p1) {
				return ErrIllegalMove
			}

			pec, err := g.brd.GetByIndex(s.ID)
			// check that piece is valid
			if err != nil || !pec.Valid() {
				return ErrPieceNil
			}

			// disallow enemy moving ally pieces
			if pec.P1 != c.p1 {
				return ErrIllegalMove
			}

			// do the order
			ret := g.brd.Move(s.ID, s.Dst)
			if ret == false {
				return ErrIllegalMove
			}

			// first off update about the move...
			err = g.UpdateAll(model.Order{
				ID:   model.OrMove,
				Data: o.Data,
			})
			if err != nil {
				return err
			}

			// then, if it's not a promotion switch turns...
			if !(s.Dst.Y == 7 || s.Dst.Y == 0) {
				// promotion
				g.SwitchTurn()
			} else {
				if pec.Kind != board.Pawn {
					g.SwitchTurn()
				}
			}

			return nil
		},
		model.OrPromote: func(c *Client, o model.Order) error {
			g := c.g
			s := &model.PromoteOrder{}

			if !c.inPromotion() {
				return ErrIllegalPromotion
			}

			err := json.Unmarshal(o.Data, s)
			// unmarshal the order
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}

			pec, err := g.brd.GetByIndex(s.ID)
			if err != nil {
				return board.ErrEmptyPiece
			}
			if pec.Kind != board.Pawn || pec.Pos.Y != board.GetEighthRank(c.p1) {
				return ErrIllegalPromotion
			}

			if s.Kind != board.Bishop && s.Kind != board.Knight && s.Kind != board.Rook && s.Kind != board.Queen {
				// only allow [bishop, knight, rook, queen]
				return ErrIllegalPromotion
			}

			pec.Kind = s.Kind
			err = g.brd.SetKind(s.ID, pec.Kind)
			if err != nil {
				return err
			}

			err = g.UpdateAll(model.Order{
				ID: model.OrPromotion,
				Parameter: model.PromotionOrder{
					ID:   s.ID,
					Kind: s.Kind,
				},
			})
			if err != nil {
				return err
			}

			g.SwitchTurn()
			return nil
		},
		model.OrCastling: func(c *Client, o model.Order) error {
			if !c.g.IsTurn(c) {
				return ErrIllegalTurn
			}
			if !c.g.canCastle[c.p1] {
				return ErrIllegalCastling
			}

			if c.inPromotion() {
				return ErrInPromotion
			}

			cast := model.CastlingOrder{}
			err := json.Unmarshal(o.Data, &cast)
			// unmarshal the order
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}

			kingid := board.GetKing(c.p1)
			rookid := int8(0)

			rid := board.GetRooks(c.p1)
			r1, r2 := rid[0], rid[1]
			src, dst := cast.Src, cast.Dst
			if (kingid != src && kingid != dst) || src != r1 && dst != r1 && src != r2 && dst != r2 {
				fmt.Println("debug 3")
				return ErrIllegalCastling
			}

			if src == r1 || dst == r1 {
				rookid = r1
			} else if src == r2 || dst == r2 {
				rookid = r2
			}

			brd := c.g.brd
			pecrook, err := brd.GetByIndex(rookid)
			if err != nil {
				return board.ErrEmptyPiece
			}
			pecking, err := brd.GetByIndex(kingid)
			if err != nil {
				return board.ErrEmptyPiece
			}

			minx, maxx := pecrook.Pos.X, pecking.Pos.X
			if minx > maxx {
				minx, maxx = maxx, minx
			}

			y := board.GetStartRow(c.p1)
			for x := minx; x < maxx; x++ {
				if x == 0 || x == 4 || x == 7 { // skip king and rook
					continue
				}

				_, _, err := brd.Get(board.Point{x, y})
				if err == nil {
					return ErrIllegalCastling
				}
			}

			if minx == 4 {
				brd.Set(rookid, board.Point{5, y})
				brd.Set(kingid, board.Point{6, y})
			} else if minx == 0 {
				brd.Set(rookid, board.Point{3, y})
				brd.Set(kingid, board.Point{2, y})
			}

			body, err := json.Marshal(model.CastlingOrder{
				Src: kingid,
				Dst: rookid,
			})
			if err != nil {
				return err
			}

			err = c.g.UpdateAll(model.Order{
				ID:   model.OrCastling,
				Data: body,
			})
			if err != nil {
				return err
			}

			c.g.SwitchTurn()

			return nil
		},
		model.OrDone: func(c *Client, o model.Order) error {
			oth := board.GetInversePlayer(c.p1)

			c.g.done = true

			return c.g.UpdateAll(model.Order{
				ID:        model.OrDone,
				Parameter: oth,
			})
		},
	}

}
