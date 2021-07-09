package rhoas

import (
	"context"
	"fmt"
	kafkamgmt "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1"
	"github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"golang.org/x/oauth2"
	"net/http"
)

func buildAuthenticatedHTTPClient(offlineToken string, clientID string, authURL string) *http.Client {
	ctx := context.Background()
	cfg := oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  fmt.Sprintf("%s/%s", authURL, "protocol/openid-connect/token"),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
	ts := cfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: offlineToken,
	})

	return oauth2.NewClient(ctx, ts)
}

func BuildKasAPIClient(offlineToken string, clientID string, authURL string) *kafkamgmtclient.APIClient {
	return kafkamgmt.NewAPIClient(&kafkamgmt.Config{
		HTTPClient: buildAuthenticatedHTTPClient(offlineToken, clientID, authURL),
	})
}
