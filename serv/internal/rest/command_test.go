package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/toms1441/chess/serv/internal/game"
)

var (
	rd1, wr1 = io.Pipe()
	rd2, wr2 = io.Pipe()

	cl1 = &game.Client{
		W: wr1,
	}

	cl2 = &game.Client{
		W: wr2,
	}

	us1 = AddClient(cl1)
	us2 = AddClient(cl2)

	gGame, _ = game.NewGame(cl1, cl2)
)

func TestCommandRequest(t *testing.T) {

	b := make([]byte, 30)
	_, err := rd1.Read(b)
	if err != nil {
		t.Fatalf("rd1.Read: %s", err.Error())
	}

	b = make([]byte, 30)
	_, err = rd2.Read(b)
	if err != nil {
		t.Fatalf("rd1.Read: %s", err.Error())
	}

	x := game.ModelCmdMessage{
		Message: "test",
	}

	byt, err := json.Marshal(x)
	if err != nil {
		t.Fatalf("json.Marshal: %s", err.Error())
	}

	cmd := game.Command{
		ID:   game.CmdMessage,
		Data: byt,
	}

	byt, err = json.Marshal(cmd)
	if err != nil {
		t.Fatalf("json.Marshal: %s", err.Error())
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(byt))

	if err != nil {
		t.Fatal(err)
	}

	handle := http.HandlerFunc(CmdHandler)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", us1.Token))

	handle.ServeHTTP(resp, req)
	hh := resp.Header()
	if hh.Get("Content-Type") != "application/json" {
		t.Fatalf("bad content type")
	}

	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatalf("ioutil.ReadAll: %s", err.Error())
	} else {
		obj := map[string]interface{}{}

		err := json.Unmarshal(p, &obj)
		if err != nil {
			t.Fatalf("json.Unmarshal: %s", err.Error())
		} else {
			err, ok := obj["error"]
			if ok {
				t.Fatalf("err: %s", err.(string))
			} else {
				t.Logf("%v", obj)

				b := make([]byte, 40)
				_, err := rd1.Read(b)
				if err != nil {
					t.Fatalf("rd1.Read: %s", err.Error())
				}
				fmt.Println(string(b))

				b = make([]byte, 40)
				_, err = rd2.Read(b)
				if err != nil {
					t.Fatalf("rd1.Read: %s", err.Error())
				}

				fmt.Println(string(b))
			}
		}
	}

}
