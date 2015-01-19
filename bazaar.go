// Bazaar is an wrapper for cafebazaar.ir purchase API
package bazaar

import (
	"encoding/json"
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
)

type config struct {
	RefreshToken string `json:"refresh_token" toml:"refresh_token"`
	AccessToken  string `json:"access_token" toml:"access_token"`
	ClientId     string `json:"client_id" toml:"client_id"`
	ClientSecret string `json:"client_secret" toml:"client_secret"`
}

type token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type Client struct {
	config config
	mu     sync.Mutex
}

// Create a new client and define it's configuration based on given argumans
func NewClient(clientId string, clientSecret string, refreshToken string) (c Client, err error) {
	c = Client{}

	c.config = config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}

	err = c.RefreshToken()
	return c, err
}

// Create a new client based on config file
func NewClientFromFile(path string) (c Client, err error) {
	c = Client{}
	if _, err = toml.DecodeFile(path, &c.config); err != nil {
		return c, err
	}

	err = c.RefreshToken()
	return c, err
}

// Refresh access token using user credentials
func (c *Client) RefreshToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	endpoint := NewEndpoint("auth")
	url := endpoint.generate()

	// Create refresh token form based with user credentials
	form := Form{
		"grant_type":    "refresh_token",
		"client_id":     c.config.ClientId,
		"client_secret": c.config.ClientSecret,
		"refresh_token": c.config.RefreshToken,
	}

	// Create request body and gather content type
	body, contentType := form.Build()

	// Initiate a new request with POST method
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return errors.New("CannotInitiateRequest")
	}

	// Set content type of the request
	req.Header.Add("Content-Type", contentType)

	// Initiate a HTTP client for sending request
	client := &http.Client{}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("CannotSendRequest")
	}

	// Close response body before function exit
	defer resp.Body.Close()

	// Check for response body
	if resp.StatusCode != http.StatusOK {
		return errors.New("InvalidCredentials")
	}

	// Decode request body into a token request
	token := token{}
	json.NewDecoder(resp.Body).Decode(&token)

	// Check if request was successful
	if token.AccessToken == "" {
		return errors.New("CannotGetAccessToken")
	}

	c.config.AccessToken = token.AccessToken
	return nil

}

// Purchase structure
type Purchase struct {
	ConsumptionState int    `json:"consumptionState"`
	PurchaseState    int    `json:"purchaseState"`
	Kind             string `json:"kind"`
	DeveloperPayload string `json:"developerPayload"`
	PurchaseTime     int    `json:"purchaseTime"`
}

// Validate a purchase
func (c *Client) PurchaseValidate(packageName string, productId string, purchaseToken string) (p Purchase, err error) {
	endpoint := NewEndpoint("validatePurchase")
	endpoint.setOption("packageName", packageName)
	endpoint.setOption("productId", productId)
	endpoint.setOption("purchaseToken", purchaseToken)
	endpoint.setAccessToken(c.config.AccessToken)

	err = c.requestTo(endpoint, &p)
	return p, err
}

// Subscription status
type Subscription struct {
	Kind                    string `json:"kind"`
	InitiationTimestampMsec int    `json:"initiationTimestampMsec"`
	ValidUntilTimestampMsec int    `json:"validUntilTimestampMsec"`
	AutoRenewing            bool   `json:"autoRenewing"`
}

// Get status of a subscription
func (c *Client) SubscriptionGet(packageName string, subscriptionId string, purchaseToken string) (s Subscription, err error) {
	endpoint := NewEndpoint("getSubscriptionStatus")
	endpoint.setOption("packageName", packageName)
	endpoint.setOption("subscriptionId", subscriptionId)
	endpoint.setOption("purchaseToken", purchaseToken)
	endpoint.setAccessToken(c.config.AccessToken)

	err = c.requestTo(endpoint, &s)
	return s, err
}

// Cancel a subscription
func (c *Client) SubscriptionCancel(packageName string, subscriptionId string, purchaseToken string) error {
	endpoint := NewEndpoint("cancelSubscription")
	endpoint.setOption("packageName", packageName)
	endpoint.setOption("subscriptionId", subscriptionId)
	endpoint.setOption("purchaseToken", purchaseToken)
	endpoint.setAccessToken(c.config.AccessToken)

	err := c.requestTo(endpoint, nil)
	return err
}

// Create a new request and loads it's output into the `output` variable(second argument)
func (c *Client) requestTo(endpoint Endpoint, output interface{}) error {
	url := endpoint.generate()
	// Initialize a new request
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("CannotSendRequest")
	}

	defer resp.Body.Close()

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return errors.New("CannotGetData")
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("CannotReadBody")
	}

	// Check access token expiration
	if string(body) == "Access token has been expired" {
		return errors.New("AccessTokenExpired")
	}

	// Check for cancel subscription
	if endpoint.Route == "cancelSubscription" && string(body) == "" {
		return nil
	}

	// Check if response was empty(this means that the server cannot find the resource!)
	if isEmpty, _ := regexp.Match("{\\s+}", body); isEmpty {
		return errors.New("CannotFindResource")
	}

	// Parse json response and load it into the `output` variable
	if err := json.Unmarshal(body, output); err != nil {
		return errors.New("Cannot parse JSON response")
	}

	return nil
}
