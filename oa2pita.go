package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/bitbucket"
)

const tokenFileName = ".bb.token"

// Authentication server will use this channel (if exists)
// to commuinicate back the authentication code it receives
// from a redirect URL a client is redirected to after authorization.
// When there's no channel registered (i.e. it's nil) the server
// ignores all incoming requests.
var registeredCodeChan chan string
var serverLock sync.Mutex

func main() {
	startAuthServer()
	tokenFile := getTokenFilePath()
	token := tryReadToken(tokenFile)
	if token == nil {
		fmt.Println("No token file exists at the moment. Trying to" +
			" get a new token...")
		token = obtainNewToken(tokenFile)
	} else if time.Now().After(token.Expiry) {
		fmt.Println("Phew! Your token has expired. Rfreshing...")
		token = refreshToken(token, tokenFile)
	}

	tokenJs, err := json.MarshalIndent(token, "", "    ")
	exitOnError(err, "Failed to marshal the token")
	fmt.Printf("Your token is: %s\n", tokenJs)
}

func startAuthServer() {
	fmt.Printf("Running auth server ...\n")
	go func() {
		http.HandleFunc("/authBitbucket", authHandler)
		http.ListenAndServe(":6162", nil)
	}()
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	serverLock.Lock()
	defer serverLock.Unlock()
	if registeredCodeChan == nil {
		return
	}

	if code, ok := r.URL.Query()["code"]; !ok || len(code) != 1 {
		registeredCodeChan <- ""
	} else {
		io.WriteString(w, "Congratulations, you're succefully authorized"+
			" with bitbucket. You can get back to your terminal now")
		registeredCodeChan <- code[0]
	}
}

func obtainNewToken(tokenFile string) *oauth2.Token {
	cfg := buildOauthConfig()
	codeChan := make(chan string)
	registerCodeChannel(codeChan)
	defer unregisterCodeChannel()

	fmt.Printf("To authorize with bitbucket please follow this url: %s\n",
		cfg.AuthCodeURL(""))
	code := <-codeChan
	if code == "" {
		reportError("Authorization failed")
	}

	token, err := cfg.Exchange(oauth2.NoContext, code)
	exitOnError(err, "Failed to get token")
	saveTokenToFile(token, tokenFile)
	return token
}

func refreshToken(staleToken *oauth2.Token, tokenFile string) *oauth2.Token {
	cfg := buildOauthConfig()
	ts := cfg.TokenSource(oauth2.NoContext, staleToken)
	token, err := ts.Token()
	exitOnError(err, "Failed to refresh the token")
	saveTokenToFile(token, tokenFile)
	return token
}

func registerCodeChannel(codeChan chan string) {
	serverLock.Lock()
	defer serverLock.Unlock()
	if registeredCodeChan == nil {
		registeredCodeChan = codeChan
	} else {
		reportError("Code channel is already registered")
	}
}

func unregisterCodeChannel() {
	serverLock.Lock()
	defer serverLock.Unlock()
	if registeredCodeChan != nil {
		close(registeredCodeChan)
		registeredCodeChan = nil
	}
}

func buildOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ask("Please enter a clinent ID: "),
		ClientSecret: ask("And a client secret: "),
		Endpoint:     bitbucket.Endpoint,
		Scopes:       []string{"project"},
	}
}

func ask(prompt string) string {
	fmt.Printf(prompt)
	var answer string
	fmt.Scanln(&answer)
	return answer
}

func getTokenFilePath() string {
	hdir, err := homedir.Dir()
	exitOnError(err, "Failed to get homedir")
	return filepath.Join(hdir, tokenFileName)
}

func tryReadToken(tokenFile string) *oauth2.Token {
	if _, err := os.Stat(tokenFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		reportError("os.Stat failed: %v", err)
	}

	bytes, err := ioutil.ReadFile(tokenFile)
	exitOnError(err, "Failed to read token")

	var token oauth2.Token
	err = json.Unmarshal(bytes, &token)
	exitOnError(err, "Failed to unmarshal token")
	return &token
}

func saveTokenToFile(token *oauth2.Token, tokenFile string) {
	bytes, err := json.Marshal(token)
	exitOnError(err, "Failed to marshal the token")
	err = ioutil.WriteFile(tokenFile, bytes, 0600)
	exitOnError(err, "Failed to write token to file %s", tokenFile)
}

func exitOnError(err error, format string, args ...interface{}) {
	if err != nil {
		reportError(format+": [%v]", append(args, err))
	}
}

func reportError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(-1)
}
