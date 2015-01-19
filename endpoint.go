package bazaar

import (
	"fmt"
)

// Create new endpoint
func NewEndpoint(route string) Endpoint {
	e := Endpoint{}
	e.setBaseUrl("https://pardakht.cafebazaar.ir")
	e.setRoute(route)
	e.Vars = map[string]string{}
	return e
}

// Endpoint structure
type Endpoint struct {
	BaseUrl      string
	Vars         map[string]string
	Route        string
	AcceessToken string
}

// Set endpoint's base url
func (e *Endpoint) setBaseUrl(url string) {
	e.BaseUrl = url
}

// Define type of route
func (e *Endpoint) setRoute(route string) {
	e.Route = route
}

// Define url option
func (e *Endpoint) setOption(key string, value string) {
	e.Vars[key] = value
}

// Set access token
func (e *Endpoint) setAccessToken(accessToken string) {
	e.AcceessToken = accessToken
}

// Generate final url
func (e *Endpoint) generate() string {
	url := e.BaseUrl

	switch e.Route {
	case "auth":
		url += "/auth/token/"
	case "validatePurchase":
		url += fmt.Sprintf("/api/validate/%s/inapp/%s/purchases/%s/", e.Vars["packageName"], e.Vars["productId"], e.Vars["purchaseToken"])
	case "getSubscriptionStatus":
		url += fmt.Sprintf("/api/applications/%s/subscriptions/%s/purchases/%s/", e.Vars["packageName"], e.Vars["subscriptionId"], e.Vars["purchaseToken"])
	case "cancelSubscription":
		url += fmt.Sprintf("/api/applications/%s/subscriptions/%s/purchases/%s/cancel/", e.Vars["packageName"], e.Vars["subscriptionId"], e.Vars["purchaseToken"])
	}

	if e.AcceessToken != "" {
		url += "?access_token=" + e.AcceessToken
	}

	return url
}
