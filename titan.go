package titan

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Location struct {
	ID         int64
	CustomerID int64
	Type       string `json:"type"` // dynamicip
	Name       string `json:"name"`
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Tag        string `json:"tag"`
	PolicyID   int64  `json:"policyid"`
	//Credential OAuthCredential
}
// OAuthCredentials is used to override the default environment variables in case a certain customer needs to be whitelisted in two places
type OAuthCredential struct {
	WebTitanBaseUrl string `json:"API_URL"`
	ConsumerKey     string `json:"CONSUMER_KEY"`
	ConsumerSecret  string `json:"CONSUMER_SECRET"`
	TokenKey        string `json:"TOKEN_KEY"`
	TokenSecret     string `json:"TOKEN_SECRET"`
}

type CustomerAccount struct {
	ID          int64
	AccountName string `json:"account_name"`
	Email       string `json:"email"`
}

// APITitan WebTitan Configuration for managed API Calls
type APITitan struct {
	Credential OAuthCredential // OAuth Credential
	Accounts []CustomerAccount
	AccountsByCode map[string]int64
	Locations []Location
	LocationsByIp map[string]Location
	LocationsByName map[string]Location
}

type UserResponse struct {
	Object         string
	ID             int64
	Created        string
	Account_name   string
	Description    string
	Timezone       string
	Email          string
	License        string
	LastLogin      string
}
// LocationsResponse is the response object returned from a getlocations get request
type UsersResponse struct {
	Object string
	Code   int64
	Count  int64
	Total  int64
	Data   []UserResponse
}

// OAuthClient returns a http client with oauth credentials properly entered
func OAuthClient(credentials OAuthCredential) *http.Client {
	config := oauth1.NewConfig(credentials.ConsumerKey, credentials.ConsumerSecret)
	token := oauth1.NewToken(credentials.TokenKey, credentials.TokenSecret)
	return config.Client(oauth1.NoContext, token)
}

// LocationResponse is a the response format of a location from a getlocations get request
type LocationResponse struct {
	Object   string
	Type     string
	Code     int64
	ID       int64
	Name     string
	PolicyID int64
	IP       string
}

// LocationsResponse is the response object returned from a getlocations get request
type LocationsResponse struct {
	Object string
	Code   int64
	Count  int64
	Total  int64
	Data   []LocationResponse
}

func (t *APITitan) GetCustomerAccounts() error {

	if len(t.Accounts) != 0 {
		return nil
	}
	endpointURL := fmt.Sprintf("%s/restapi/users", t.Credential.WebTitanBaseUrl)
	client := OAuthClient(t.Credential)
	log.Info("Retrieving WebTitan UserAccounts", zap.String("ENDPOINT_URL", endpointURL))

	response, err := client.Get(endpointURL)
	if err != nil {
		log.Error("Error get GetListUserAccounts", zap.Error(err))
		return err
	}
	//header := response.Request.Header.Get("authorization")
	//if err == nil {
	//	l.Info("", zap.String("REQUEST_HEADER", header))
	//}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("Error body", zap.Error(err))
		return err
	}

	usersResponse := UsersResponse{}

	err = json.Unmarshal(body, &usersResponse)
	if err != nil {
		log.Error("Error json.Unmarshal", zap.Error(err))
		return err
	}

	t.AccountsByCode = make(map[string]int64)
	for _, l := range usersResponse.Data {
		t.AccountsByCode[l.Account_name] = l.ID
		t.Accounts = append(t.Accounts, CustomerAccount{ID: l.ID, AccountName: l.Account_name, Email: l.Email})
	}

	return nil
}

func (t *APITitan) GetLocationById(CustomerId int64) (LocationsResponse, error) {

	endpointURL := fmt.Sprintf("%s/restapi/users/%d/locations/dynamicip", t.Credential.WebTitanBaseUrl, CustomerId)
	client := OAuthClient(t.Credential)
	log.Info("Retrieving WebTitan Locations", zap.String("ENDPOINT_URL", endpointURL))

	response, err := client.Get(endpointURL)
	if err != nil {
		log.Error("Error Retrieving WebTitan Locations", zap.Error(err))
		return LocationsResponse{}, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("Error ioutil.ReadAll(response.Body)", zap.Error(err))
		return LocationsResponse{}, err
	}

	locationsRespone := LocationsResponse{}
	err = json.Unmarshal(body, &locationsRespone)
	if err != nil {
		log.Error("Error json.Unmarshal", zap.Error(err))
		return LocationsResponse{} ,err
	}
	return locationsRespone, nil
}

func (t *APITitan) GetLocations() error {

	err := t.GetCustomerAccounts()
	if err != nil {
		log.Error("Error get GetListUserAccounts", zap.Error(err))
		return err
	}

	t.LocationsByIp = make(map[string]Location)
	t.LocationsByName = make(map[string]Location)

	for _, a := range t.Accounts {
		locationsRespone, err := t.GetLocationById(a.ID)
		if err != nil {
			log.Error("Error json.Unmarshal", zap.Error(err))
			continue
		}

		for _, l := range locationsRespone.Data {
			location := Location{ID: l.ID, Name: l.Name, IP: l.IP, CustomerID: a.ID}
			t.Locations = append(t.Locations, location)
			t.LocationsByIp[strings.TrimSpace(l.IP)] = location
			t.LocationsByName[strings.TrimSpace(l.Name)] = location
		}
	}

	return nil
}

func (t *APITitan) UpdateLocation(location Location) error {

	v := url.Values{}
	v.Set("ip", location.IP)
	v.Set("name", location.Name)

	var endpointURL string

	if location.ID == 0 {
		endpointURL = fmt.Sprintf("%s/restapi/users/%d/locations/dynamicip",
			t.Credential.WebTitanBaseUrl, location.CustomerID)
	} else {
		endpointURL = fmt.Sprintf("%s/restapi/users/%d/locations/dynamicip/%d",
			t.Credential.WebTitanBaseUrl, location.CustomerID, location.ID)
	}

	client := OAuthClient(t.Credential)
	response, err := client.Post(endpointURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(v.Encode()))

	if err != nil {
		return err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("endpointURL: %s, name: %s, ip: %s, status: %s, body: %s",
			endpointURL, location.Name, location.IP, response.Status, string(body))
	}

	log.Info("Updated location",
		zap.String("HTTP_STATUS", response.Status),
		zap.String("HTTP_BODY", string(body)),
	)
	return nil
}

func (t *APITitan) DeleteLocation(location Location) error {
	if location.CustomerID == 0 || location.ID == 0 || location.IP != "" {
		return nil
	}

	endpointURL := fmt.Sprintf("%s/restapi/users/%d/locations/dynamicip/%d",
		t.Credential.WebTitanBaseUrl, location.CustomerID, location.ID)

	client := OAuthClient(t.Credential)

	req, err := http.NewRequest("DELETE", endpointURL, nil)
	if err != nil {
		return err
	}
	// Fetch Request
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("id: %d, locationId: %d, status: %s, body: %s",
			location.CustomerID, location.ID, response.Status, string(body))
	}

	log.Info("Delete location",
		zap.String("HTTP_STATUS", response.Status),
		zap.String("HTTP_BODY", string(body)),
	)
	return nil
}
