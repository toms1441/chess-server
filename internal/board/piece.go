package board

import (
	"fmt"
)

const (
	Empty uint8 = iota
	// Pawn
	// Moves forward(0, 6 -> 0, 5) if P1, other moves backward(0, 1 -> 0, 2)
	Pawn
	// Bishop
	// Moves diagonally
	Bishop
	// Knight
	// Moves with [2, 1] or [1, 2]
	Knight
	// Rook
	// Moves horizontally, vertically.
	Rook
	// Queen
	// Moves horizontally, vertically, diagonally, and within square.
	Queen
	// King
	// Moves within square.
	King
)

type Piece struct {
	// P1 is set to true, if it's the first player. if it's the second player then it's set to false.
	// Note: Player 1 always starts in Y {6, 7}
	P1 bool `json:"p1"`
	// Kind is the type of piece, since type is keyword in most langauges..
	Kind uint8 `json:"kind"`
	// Pos where the piece stands..
	Pos Point `json:"pos"`
}

// Valid returns true if the piece is true. This basically checks if the type is in bound, if the point valid
func (p Piece) Valid() bool {
	if !p.Pos.Valid() {
		return false
	}

	if p.Kind >= Pawn && p.Kind <= King {
		return true
	}

	return false
}

// ShortString produces a one-character string to represent the piece. Used for debugging.
func (p Piece) ShortString() string {
	if p.Kind == Empty {
		return " "
	} else if p.Kind == Knight {
		return "N"
	}

	name := p.Name()

	return name[:1]
}

// Name returns the name type for the piece
func (p Piece) Name() string {
	strings := map[uint8]string{
		Empty:  "Empty",
		Pawn:   "Pawn",
		Bishop: "Bishop",
		Knight: "Knight",
		Rook:   "Rook",
		Queen:  "Queen",
		King:   "King",
	}

	return strings[p.Kind]
}

// String representation of the piece's [type, number, position]
func (p Piece) String() string {
	num := 0
	if p.P1 {
		num = 1
	}

	return fmt.Sprintf("%s/%s/%d", p.Name(), p.Pos.String(), num)
}

// Possib returns all.Possible moves from piece's.Position.
func (p Piece) Possib() Points {
	src := p.Pos
	// i.e starting point
	switch p.Kind {
	// Only horizontally, can't move back
	// 2 points at start, 1 point after that
	case Pawn:
		ps := Points{}
		if p.P1 {
			ps = src.Forward()
		} else {
			ps = src.Backward()
		}
		if src.Y == 1 || src.Y == 6 {
			if p.P1 {
				ps.Insert(Point{X: src.X, Y: src.Y - 2})
			} else {
				ps.Insert(Point{X: src.X, Y: src.Y + 2})
			}
		}

		ps.Clean()
		return ps
	// Only diagonal
	case Bishop:
		return src.Diagonal()
	// Move within [2, 1] or [1, 2]
	case Knight:
		return src.Knight()
	// horizontal or vertical
	case Rook:
		return src.Horizontal().Merge(src.Vertical())
	// move within square or diagonal or horizontal or vertical
	case Queen:
		return src.Horizontal().
			Merge(src.Vertical()).
			//Merge(src.Square()). no need vertical|horizontal|diagonal includes square
			Merge(src.Diagonal())
	// move within square
	case King:
		return src.Square()
	}

	return Points{}
}

// CanGo does validation for the piece. Each piece has it's own rules.
func (p Piece) CanGo(dst Point) bool {
	if dst.Equal(p.Pos) || !dst.Valid() {
		return false
	}

	return p.Possib().In(dst)
}
