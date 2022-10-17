package api

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	svcacctmgmtclient "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	"github.com/redhat-developer/terraform-provider-rhoas/rhoas/localize"
	"net/http"
)

type Factory interface {
	KafkaMgmt() kafkamgmtclient.DefaultApi
	ServiceAccountMgmt() svcacctmgmtclient.ServiceAccountsApi
	KafkaAdmin(ctx *context.Context, instanceID string) (*kafkainstanceclient.APIClient, *kafkamgmtclient.KafkaRequest, error)
	HTTPClient() *http.Client
	Localizer() localize.Localizer
}
