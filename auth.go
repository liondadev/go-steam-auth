package gosteamauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ErrInvalidAuthRequest is returned by ValidateCallback when the auth attempt is invalid, as stated
// by steam's auth servers. This may be due to the user attempting to impersonate someone else.
var ErrInvalidAuthRequest = errors.New("invalid authentication attempt")

// ErrNoData is returned by GetSteamUser if steam doesn't return any data about the provided steamid64.
var ErrNoData = errors.New("steam did not return any data about the provided user")

// SteamAuther provides methods to handle authentication for steam users.
type SteamAuther struct {
	// apiKey is the Steam Web API Key.
	// These are provisioned at https://steamcommunity.com/dev/apikey
	apiKey string

	// realm is the openid2 realm.
	// This should be the base URL of your web application, in most scenarios. For example,
	realm string
}

// New returns a new SteamAuther with the provided options.
// apiKey is the steam web api key. realm is the openid 2 realm (typically the base url to your application (ex. http://localhost:8080))
func New(apiKey, realm string) *SteamAuther {
	return &SteamAuther{
		apiKey: apiKey,
		realm:  realm,
	}
}

// OpenIdLoginUrl is from https://steamcommunity.com/openid/, hardcoded because it's unlikely this will ever change.
const OpenIdLoginUrl = "https://steamcommunity.com/openid/login"

// GetAuthUrl generates an OpenID2 URL to redirect the user to in order to start the authentication process.
// The user should be redirected here when you want to start the OAuth2 flow.
// returnUrl is the url to return the user to once they've signed in. See ValidateCallback for what to do in that handler.
func (sa *SteamAuther) GetAuthUrl(returnUrl string) (string, error) {
	u, err := url.Parse(OpenIdLoginUrl)
	if err != nil {
		return "", fmt.Errorf("get redirect url (returnUrl=\"%s\"): %w", returnUrl, err)
	}

	// Setup openid query params
	q := u.Query()
	q.Set("openid.ns", "http://specs.openid.net/auth/2.0")                           // this is an openid 2.0 request
	q.Set("openid.mode", "checkid_setup")                                            // we're planning on verifying the authentication request ourself
	q.Set("openid.realm", sa.realm)                                                  // we're doing the authentication
	q.Set("openid.return_to", returnUrl)                                             // return to our webapp
	q.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select") // the user hasn't asserted who they are yet
	q.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")   // the user hasn't asserted who they are yet
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ValidateCallback is used to validate the callback at the end of an openid2 flow. This returns the steamid64 or an error.
// This is used in the route handler that's at the returnUrl given at the start of the flow.
// The vals correspond to the URL query parameters in the callback request.
func (sa *SteamAuther) ValidateCallback(vals url.Values) (string, error) {
	// To validate the callback, we just take the raw params provided by the user and call back
	// to steam to make sure everything is valid. This is required to make sure we're not getting epically pranked by
	// someone trying to impersonate someone else.

	if vals.Get("openid.mode") != "id_res" {
		return "", fmt.Errorf("the openid.mode was not expected. got=%x, expected=id_res", vals.Get("openid.mode"))
	}

	vals.Set("openid.mode", "check_authentication") // tell steam we're trying to validate an auth response
	res, err := http.Post(OpenIdLoginUrl, "application/x-www-form-urlencoded", bytes.NewReader([]byte(vals.Encode())))
	if err != nil {
		return "", fmt.Errorf("validate callback: failed making validation request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("validate callback: read all bytes: %w", err)
	}

	if !strings.Contains(string(bodyBytes), "is_valid:true") {
		return "", ErrInvalidAuthRequest
	}

	// The callback is ok, so we need to split out the steamid
	p := strings.Split(vals.Get("openid.claimed_id"), "/")
	return p[len(p)-1], nil
}

// GetSteamUser gets the steamid user with the steamid64 provided and returns some basic information about them.
// This is useful to check after using ValidateCallback to get info about the user that's being authenticated.
// It's a good idea to copy and store this somewhere else to prevent being dependent on steam for every request to
// your website.
func (sa *SteamAuther) GetSteamUser(steamid64 string) (*SteamUser, error) {
	// First, we need to build the URL that we'll be making the request to.
	u, err := url.Parse("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002")
	if err != nil {
		return nil, fmt.Errorf("get steam user (%s): parse api url: %w", steamid64, err)
	}

	q := u.Query()
	q.Set("key", sa.apiKey)
	q.Set("steamids", steamid64)
	u.RawQuery = q.Encode() // I can't believe this is required...

	reqUrl := u.String()

	// Now we need to *do* the request :)
	res, err := http.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("get steam user (%s): make get request: %w", steamid64, err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("get steam user (%s): status code is not 200 (%s)", steamid64, res.Status)
	}

	var data struct {
		Response struct {
			Players []SteamUser `json:"players"`
		} `json:"response"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("get steam user (%s): decode response body: %w", steamid64, err)
	}

	if len(data.Response.Players) < 1 {
		return nil, ErrNoData
	}

	return &data.Response.Players[0], nil
}
