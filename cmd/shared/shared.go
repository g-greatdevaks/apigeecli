package shared

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/lestrrat/go-jwx/jwa"
	"github.com/lestrrat/go-jwx/jwt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://apigee.googleapis.com/v1/organizations/"

// Arguements is the base struct to hold all command arguments
type Arguments struct {
	Verbose        bool
	Org            string
	Env            string
	Token          string
	ServiceAccount string
}

var RootArgs = Arguments{}

//log levels, default is error
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

var LogInfo = false
var skipCheck = false
var skipCache = false

// Structure to hold OAuth response
type OAuthAccessToken struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
}

const access_token_file = ".access_token"

//Init function initializes the logger objects
func Init() {

	var infoHandle = ioutil.Discard

	if LogInfo {
		infoHandle = os.Stdout
	}

	warningHandle := os.Stdout
	errorHandle := os.Stdout

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func PostHttpOctet(url string, proxyName string) error {

	file, _ := os.Open(proxyName)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("proxy",proxyName)
	if err != nil {
		Error.Fatalln("Error writing multi-part:\n", err)
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		Error.Fatalln("Error copying multi-part:\n", err)
		return err
	}

	err = writer.Close()
	if err != nil {
		Error.Fatalln("Error closing multi-part:\n", err)
		return err
	}
	client := &http.Client{}

	Info.Println("Connecting to : ", url)
	req, err := http.NewRequest("POST", url, body)

	Info.Println("Setting token : ", RootArgs.Token)
	req.Header.Add("Authorization", "Bearer "+ RootArgs.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		Error.Fatalln("Error connecting:\n", err)
		return err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Error.Fatalln("Error in response:\n", err)
			return err
		} else if resp.StatusCode != 200 {
			Error.Fatalln("Error in response:\n", string(body))
			return errors.New("Error in response")
		} else {
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, body, "", "\t")
			if err != nil {
				Error.Fatalln("Error parsing response:\n", err)
				return err
			}
			fmt.Println(string(prettyJSON.Bytes()))
			return nil
		}
	}
}

func HttpClient(params ...string) error {
	
	var req *http.Request
	var err error

	client := &http.Client{}
	Info.Println("Connecting to : ", params[0])

	if len(params) == 2 {
		req, err = http.NewRequest("POST", params[0], bytes.NewBuffer([]byte(params[1])))
	} else if len(params) == 1 {
		req, err = http.NewRequest("GET", params[0], nil)		
	} else {
		return errors.New("Incorrect parameters to invoke the method")
	}

	Info.Println("Setting token : ", RootArgs.Token)
	req.Header.Add("Authorization", "Bearer "+ RootArgs.Token)

	resp, err := client.Do(req)

	if err != nil {
		Error.Fatalln("Error connecting:\n", err)
		return err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Error.Fatalln("Error in response:\n", err)
			return err
		} else if resp.StatusCode != 200 {
			Error.Fatalln("Error in response:\n", string(body))
			return errors.New("Error in response")
		} else {
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, body, "", "\t")
			if err != nil {
				Error.Fatalln("Error parsing response:\n", err)
				return err
			}
			fmt.Println(string(prettyJSON.Bytes()))
			return nil
		}
	}	
}

func getPrivateKey() (interface{}, error) {
	pemPrivateKey := fmt.Sprintf("%v", viper.Get("private_key"))
	block, _ := pem.Decode([]byte(pemPrivateKey))
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		Error.Fatalln("Error parsing Private Key:\n", err)
		return nil, err
	} else {
		return privKey, nil
	}
}

func generateJWT() (string, error) {

	const aud = "https://www.googleapis.com/oauth2/v4/token"
	const scope = "https://www.googleapis.com/auth/cloud-platform"

	privKey, err := getPrivateKey()

	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.New()

	token.Set(jwt.AudienceKey, aud)
	token.Set(jwt.IssuerKey, viper.Get("client_email"))
	token.Set("scope", scope)
	token.Set(jwt.IssuedAtKey, now.Unix())
	token.Set(jwt.ExpirationKey, now.Unix())

	payload, err := token.Sign(jwa.RS256, privKey)
	if err != nil {
		Error.Fatalln("Error parsing Private Key:\n", err)
		return "", err
	} else {
		Info.Println("jwt token : ", string(payload))
		return string(payload), nil
	}
}

func GenerateAccessToken() (string, error) {

	const token_endpoint = "https://www.googleapis.com/oauth2/v4/token"
	const grant_type = "urn:ietf:params:oauth:grant-type:jwt-bearer"

	token, err := generateJWT()

	if err != nil {
		return "", nil
	}

	form := url.Values{}
	form.Add("grant_type", grant_type)
	form.Add("assertion", token)

	client := &http.Client{}
	req, err := http.NewRequest("POST", token_endpoint, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	resp, err := client.Do(req)

	if err != nil {
		Error.Fatalln("Failed to generate oauth token: \n", err)
		return "", err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			Error.Fatalln("Error in response: \n", string(bodyBytes))
			return "", errors.New("Error in response")
		} else {
			decoder := json.NewDecoder(resp.Body)
			accessToken := OAuthAccessToken{}
			if err := decoder.Decode(&accessToken); err != nil {
				Error.Fatalln("Error in response: \n", err)
				return "", errors.New("Error in response")
			} else {
				Info.Println("access token : ", accessToken)
				RootArgs.Token = accessToken.AccessToken
				writeAccessToken()
				return accessToken.AccessToken, nil
			}
		}
	}
}

func readAccessToken() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(path.Join(usr.HomeDir,access_token_file))
	if err != nil {
		Info.Println("Cached access token was not found")
		return err
	} else {
		Info.Println("Using cached access token: ", string(content))
		RootArgs.Token = string(content)
		return nil
	}
}

func writeAccessToken() error {

	if skipCache {
		return nil
	}
	usr, err := user.Current()
	if err != nil {	
		Warning.Println(err)
	} else {
		Info.Println("Cache access token: ", RootArgs.Token)
		err = ioutil.WriteFile(path.Join(usr.HomeDir,access_token_file), []byte(RootArgs.Token), 0644)
	}
	return err
}

func checkAccessToken() bool {

	if skipCheck {
		Warning.Println("skipping token validity")
		return true
	}

	const tokenInfo = "https://www.googleapis.com/oauth2/v1/tokeninfo"
	u, _ := url.Parse(tokenInfo)
	q := u.Query()
	q.Set("access_token", RootArgs.Token)	
	u.RawQuery = q.Encode()

	client := &http.Client{}

	Info.Println("Connecting to : ", u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	resp, err := client.Do(req)
	if err != nil {
		Error.Fatalln("Error connecting to token endpoint:\n", err)
		return false
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Error.Fatalln("Token info error:\n", err)
			return false
		} else if resp.StatusCode != 200 {
			Error.Fatalln("Token expired:\n", string(body))
			return false
		} else {
			Info.Println("Response: ", string(body))
			Info.Println("Reusing the cached token: ", RootArgs.Token)
			return true
		}
	}
}

func SetAccessToken () error {

	if RootArgs.Token == "" && RootArgs.ServiceAccount == "" {
		err := readAccessToken() //try to read from config
		if err != nil {
			return fmt.Errorf("Either token or service account must be provided")
		} else {
			if checkAccessToken() { //check if the token is still valid
				return nil
			} else {
				return fmt.Errorf("Token expired: request a new access token or pass the service account")
			}			
		}
	} else {
		if RootArgs.ServiceAccount != "" {
			viper.SetConfigFile(RootArgs.ServiceAccount)
			err := viper.ReadInConfig() // Find and read the config file
			if err != nil {             // Handle errors reading the config file
				return fmt.Errorf("Fatal error config file: %s \n", err)
			} else {
				if viper.Get("private_key") == "" {
					return fmt.Errorf("Fatal error: Private key missing in the service account")
				}
				if viper.Get("client_email") == "" {
					return fmt.Errorf("Fatal error: client email missing in the service account")
				}
				_, err = GenerateAccessToken()
				if err != nil {
					return fmt.Errorf("Fatal error generating access token: %s \n", err)
				} else {
					return nil
				}
			}
		} else {
			//a token was passed, cache it
			if checkAccessToken() {
				writeAccessToken()
				return nil
			} else {
				return fmt.Errorf("Token expired: request a new access token or pass the service account")
			}
		}		
	}
}