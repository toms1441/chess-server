package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/toms1441/chess-server/internal/model"
	"github.com/toms1441/chess-server/internal/model/discord"
	"github.com/toms1441/chess-server/internal/model/github"
	"github.com/toms1441/chess-server/internal/model/google"
	"github.com/toms1441/chess-server/internal/rest"
	"github.com/toms1441/chess-server/internal/rest/auth"
)

const apiver = "v1"

func main() {
	if debug != "no" && debug != "normal" {
		go debug_game()
	}

	rout := mux.NewRouter()
	api := rout.PathPrefix("/api/" + apiver).Subrouter()
	{ // api routes

		api.HandleFunc("/cmd", rest.CmdHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/invite", rest.InviteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/accept", rest.AcceptInviteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/ws", rest.WebsocketHandler).Methods("GET", "OPTIONS")
		api.HandleFunc("/avali", rest.GetAvaliableUsersHandler).Methods("GET", "OPTIONS")
		api.HandleFunc("/possib", rest.PossibHandler).Methods("POST", "OPTIONS")

		// is connected to ws?
		api.HandleFunc("/connected", func(w http.ResponseWriter, r *http.Request) {
			_, err := rest.GetUser(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			w.Write(nil)
		}).Methods("GET", "OPTIONS")
		// is logged in?
		api.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
			mo := auth.Identify(r)
			if mo == nil {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			w.Write(nil)
		}).Methods("GET", "OPTIONS")

		watchable := api.PathPrefix("/watchable/").Subrouter()
		{
			watchable.HandleFunc("/list", rest.WatchableListHandler).Methods("GET", "OPTIONS")
			watchable.HandleFunc("/join", rest.WatchableJoinHandler).Methods("POST", "OPTIONS")
			watchable.HandleFunc("/leave", rest.WatchableLeaveHandler).Methods()
		}
	}

	addplatform := func(platform string, fn func(model.OAuth2Config) auth.Config) {
		id := os.Getenv(strings.ToUpper(platform + "_CLIENT_ID"))
		secret := os.Getenv(strings.ToUpper(platform + "_CLIENT_SECRET"))
		redirect := os.Getenv(strings.ToUpper(platform + "_REDIRECT"))

		if len(id) > 0 && len(secret) > 0 && len(redirect) > 0 {
			router := api.PathPrefix("/" + platform).Subrouter()
			config := fn(model.OAuth2Config{
				ClientID:     id,
				ClientSecret: secret,
				Redirect:     redirect,
			})
			auth.AddRoutes(config, router)
		}
	}
	addplatform("discord", discord.NewAuthConfig)
	addplatform("google", google.NewAuthConfig)
	addplatform("github", github.NewAuthConfig)

	var proto string
	var port string
	if debug != "no" {
		proto = "tcp"
		port = ":8080"
	} else {
		proto = "unix"
		port = "http.sock"

		os.Remove(port)
	}

	color.New(color.FgBlue).Println("Listening on", port)

	listen, err := net.Listen(proto, port)
	if err != nil {
		panic(err)
	}

	if proto == "unix" {
		os.Chmod(port, 0777)
	}

	defer listen.Close()

	http.Serve(listen, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &rest.Context{
			ResponseWriter: w,
		}

		method := color.New(color.BgMagenta, color.Bold).Sprint(" " + r.Method + " ")
		path := color.New(color.BgBlue).Sprint(" " + r.URL.Path + " ")

		if debug != "no" {
			ctx.Header().Add("Access-Control-Allow-Origin", "*")
			ctx.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, Cookie")
			ctx.Header().Add("Access-Control-Allow-Methods", "GET, POST")
			ctx.Header().Add("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == "OPTIONS" {
			ctx.WriteHeader(http.StatusOK)
		} else {
			rout.ServeHTTP(ctx, r)
		}

		code := ""
		sta := ctx.GetStatus()
		if sta <= 299 && sta >= 200 {
			code = color.New(color.BgGreen, color.Bold).Sprintf(" %d ", sta)
		} else if sta >= 400 && sta <= 499 {
			code = color.New(color.BgYellow, color.Bold).Sprintf(" %d ", sta)
		} else if sta >= 500 && sta <= 511 {
			code = color.New(color.BgRed, color.Bold).Sprintf(" %d ", sta)
		} else {
			code = color.New(color.Reset).Sprintf(" %d ", sta)
		}

		fmt.Printf("%s%s%s\n", method, path, code)
	}))
}
