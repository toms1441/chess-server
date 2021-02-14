package rest

import (
	"net/http"

	"github.com/toms1441/chess/serv/internal/game"
)

func CmdHandler(w http.ResponseWriter, r *http.Request) {
	cl, err := GetUser(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err)
		return
	}

	cmd := game.Command{}
	err = bindJSON(r, &cmd)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	g := cl.Game()
	if g == nil {
		respondError(w, http.StatusNotFound, game.ErrGameNil)
		return
	}

	err = cl.Do(cmd)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
