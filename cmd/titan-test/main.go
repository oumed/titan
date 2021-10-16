package main

import (
	"fmt"
	t "github.com/oumed/titan"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func main() {
	credential := t.OAuthCredential {
		WebTitanBaseUrl: "",
		ConsumerKey: "",
		ConsumerSecret: "",
		TokenKey: "",
		TokenSecret: "",
	}
	fmt.Println(credential)

	apiTitan := t.APITitan{Credential: credential}
	err := apiTitan.GetLocations()
	if err != nil {
		log.Error("failed to Titan locations", zap.Error(err))
		return
	}

	fmt.Println(apiTitan.Accounts)
	fmt.Println("############################################")
	fmt.Println(apiTitan.LocationsByName)
	fmt.Println("############################################")
	fmt.Println(apiTitan.LocationsByIp)

}