package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/j4rv/golang-stuff/cah"
)

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func processLogin(w http.ResponseWriter, req *http.Request) {
	var payload loginPayload
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		http.Error(w, "Misconstructed payload", http.StatusBadRequest)
		return
	}
	u, ok := usecase.User.Login(payload.Username, payload.Password)
	if !ok {
		log.Printf("%s tried to login using user '%s'", req.RemoteAddr, payload.Username)
		http.Error(w, "The username and password you entered did not match our records.", http.StatusForbidden)
		return
	}
	session, err := cookies.Get(req, sessionid)
	session.Values[userid] = u.ID
	session.Save(req, w)
	log.Printf("User %s with id %d just logged in!", u.Username, u.ID)
	// everything ok, back to index with your brand new session!
	http.Redirect(w, req, "/", http.StatusFound)
}

func processRegister(w http.ResponseWriter, req *http.Request) {
	var payload registerPayload
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		http.Error(w, "Misconstructed payload", http.StatusBadRequest)
		return
	}
	u, err := usecase.User.Register(payload.Username, payload.Password)
	if err != nil {
		log.Printf("%s tried to register using user '%s'", req.RemoteAddr, payload.Username)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, err := cookies.Get(req, sessionid)
	session.Values[userid] = u.ID
	session.Save(req, w)
	log.Printf("User %s with id %d just registered!", u.Username, u.ID)
	// everything ok, back to index with your brand new session!
	http.Redirect(w, req, "/", http.StatusFound)
}

func processLogout(w http.ResponseWriter, req *http.Request) {
	session, err := cookies.Get(req, sessionid)
	if err != nil {
		http.Error(w, "There was a problem while getting the session cookie", http.StatusInternalServerError)
	}
	session.Values = make(map[interface{}]interface{})
	session.Save(req, w)
	http.Redirect(w, req, "/", http.StatusFound)
}

func validCookie(w http.ResponseWriter, req *http.Request) {
	_, err := userFromSession(req)
	ok := strconv.FormatBool(err == nil)
	w.Write([]byte(ok))
}

/*
	SESSIONS STUFF
*/

var cookies *sessions.CookieStore

const sessionid = "session_token"
const userid = "user_id"

func init() {
	skey := os.Getenv("SESSION_KEY")
	if skey == "" {
		panic("Please set SESSION_KEY environment variable; it is needed to have secure cookies")
	}
	cookies = sessions.NewCookieStore([]byte(skey))
}

func userFromSession(r *http.Request) (cah.User, error) {
	session, err := cookies.Get(r, sessionid)
	if err != nil {
		return cah.User{}, err
	}
	val, ok := session.Values[userid]
	if !ok {
		return cah.User{}, fmt.Errorf("Tried to get user from session without an id")
	}
	id, ok := val.(int)
	if !ok {
		log.Printf("Session with non int id value: '%v'", session.Values)
		return cah.User{}, fmt.Errorf("Session with non int id value")
	}
	u, ok := usecase.User.ByID(id)
	if !ok {
		return u, fmt.Errorf("No user found with ID %d", id)
	}
	return u, nil
}
