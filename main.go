package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

var (
	conf *oauth2.Config
	ctx  context.Context
)

const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`

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

	conf := &oauth2.Config{
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

	//test := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)

	//println(test)

}

func getAuthCode(config *oauth2.Config) {

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	fmt.Println("Trying to get token from prompt")
	getTokenFromPrompt(config, authURL)

	//var code string
	//if _, err := fmt.Scan(&code); err != nil {
	//	log.Fatalf("Unable to read authorization code %v", err)
	///}

}

func startWebServer() (codeCh chan string, err error) {
	listener, err := net.Listen("tcp", "localhost:8090")
	if err != nil {
		return nil, err
	}
	codeCh = make(chan string)

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		codeCh <- code // send code to OAuth flow
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Received code: %v\r\nYou can now safely close this browser window.", code)
	}))

	return codeCh, nil
}

// getTokenFromPrompt uses Config to request a Token and prompts the user
// to enter the token on the command line. It returns the retrieved Token.
func getTokenFromPrompt(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	var code string
	fmt.Printf("Go to the following link in your browser. After completing "+
		"the authorization flow, enter the authorization code on the command "+
		"line: \n%v\n", authURL)

	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}
	fmt.Println("FINAL: ", authURL)
	return exchangeToken(config, code)
}

// Exchange the authorization code for an access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token %v", err)
	}
	println("THE TOKEN: ", tok)
	return tok, nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {

	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect
	// URL.
	code := queryParts["code"][0]
	println("11111111111111")
	log.Printf("code: %s\n", code)
	println("2222222222222")
	// Exchange will do the handshake to retrieve the initial access token.
	println("CTX: ", ctx)
	println("CODE: ", code)
	tok, err := conf.Exchange(oauth2.NoContext, code)
	println("3333333333333")

	if err != nil {
		println("MAJOR ERROR....!")
		log.Fatal(err)
	}
	println("444444444444")
	log.Printf("Token: %s", tok)
}

func main() {
	ctx = context.Background()
	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)
	http.HandleFunc("/test", beginAuth)
	http.HandleFunc("/oauth/callback", callbackHandler)
	http.ListenAndServe(":9999", nil)
}
