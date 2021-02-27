package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kjk/betterguid"
	"github.com/thanhpk/randstr"
	"github.com/toms1441/chess-server/internal/game"
)

type User struct {
	Token    string `validate:"required" json:"string"`
	PublicID string `json:"id"`
	invite   map[string]*User
	cl       *game.Client
}

var users = map[string]*User{}

const (
	InviteLifespan = time.Second * 10
)

func AddClient(c *game.Client) *User {
	id := betterguid.New()
	us := &User{
		Token:  id,
		invite: map[string]*User{},
		cl:     c,
	}

	us.PublicID = randstr.String(4)

	users[id] = us
	return us
}

func GetUser(r *http.Request) (*User, error) {
	str := r.Header.Get("Authorization")
	str = strings.ReplaceAll(str, "Bearer ", "")
	cl, ok := users[str]

	if !ok {
		return nil, game.ErrClientNil
	}

	return cl, nil
}

func (u *User) Client() *game.Client {
	return u.cl
}

func (u *User) Delete() {
	u.cl = nil
	u.invite = nil
	delete(users, u.Token)
}

func (u *User) Valid() bool {
	if u.Client() == nil {
		return false
	}
	x, ok := users[u.Token]
	if !ok {
		return false
	}
	if x != u {
		u.Delete()
		return false
	}

	return true
}

func (u *User) Invite(tok string, lifespan time.Duration) error {
	// make sure panic don't happen
	if !u.Valid() {
		return game.ErrClientNil
	}
	if u.Client().Game() != nil {
		return game.ErrGameIsNotNil
	}

	var vs *User
	for _, v := range users {
		if v.PublicID == tok {
			vs = v
			break
		}
	}

	if vs == nil || !vs.Valid() {
		return game.ErrClientNil
	}
	if vs.Client().Game() != nil {
		return game.ErrGameIsNotNil
	}

	// lamo you though you were slick
	if vs == u {
		return game.ErrClientNil
	}

	id := randstr.String(4)
	param := game.ModelUpdateInvite{
		ID: id,
	}

	body, err := json.Marshal(param)
	if err != nil {
		return err
	}

	// u invited vs
	vs.invite[id] = u
	gu := game.Update{
		ID:   game.UpdateInvite,
		Data: body,
	}
	send, err := json.Marshal(gu)
	if err != nil {
		return err
	}

	// delete after X amount of time
	go func(vs *User, id string) {
		<-time.After(lifespan)
		delete(vs.invite, id)
	}(vs, id)

	vs.Client().W.Write(send)

	return nil
}

// AcceptInvite accepts the invite from the user.
func (u *User) AcceptInvite(tok string) error {
	vs, ok := u.invite[tok]
	if !ok {
		return ErrInvalidInvite
	}

	if vs == nil || vs.Client() == nil {
		return game.ErrClientNil
	}

	_, err := game.NewGame(u.Client(), vs.Client())
	if err != nil {
		return err
	}

	u.invite = map[string]*User{}

	return nil
}
