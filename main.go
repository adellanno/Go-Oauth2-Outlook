package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

var (
	conf *oauth2.Config
	ctx  context.Context
)

// Credentials here
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

func beginAuth(w http.ResponseWriter, r *http.Request) {
	var c Credentials
	file, err := ioutil.ReadFile("./creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &c)

	conf = &oauth2.Config{
		ClientID:     c.Cid,
		ClientSecret: c.Csecret,
		RedirectURL:  "http://127.0.0.1:9999/oauth/callback",
		Scopes: []string{
			"https://outlook.office.com/calendars.read", // these are permissions set in the application in web
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	getAuthCode(conf)
}

func getAuthCode(config *oauth2.Config) {

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	fmt.Println("Trying to get token from prompt")

	getTokenFromPrompt(config, authURL)
}

// getTokenFromPrompt uses Config to request a Token and prompts the user
// to enter the token on the command line. It returns the retrieved Token.
func getTokenFromPrompt(config *oauth2.Config, authURL string) {
	var code string
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}
	fmt.Println("FINAL: ", authURL)
}

// Exchange the authorization code for an access token
func callbackHandler(w http.ResponseWriter, r *http.Request) {

	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect
	// URL.
	code := queryParts["code"][0]

	log.Printf("code: %s\n", code)

	tok, err := conf.Exchange(oauth2.NoContext, code)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Token: %s", tok)
}

func main() {
	http.HandleFunc("/auth", beginAuth)
	http.HandleFunc("/oauth/callback", callbackHandler)
	http.ListenAndServe(":9999", nil)
}
