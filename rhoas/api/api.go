package api

import (
	"context"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1/client"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
	"net/http"
)

type API interface {
	KafkaMgmt() kafkamgmtclient.DefaultApi
	ServiceAccountMgmt() serviceAccounts.ServiceAccountsApi
	KafkaAdmin(ctx *context.Context, instanceID string) (*kafkainstanceclient.APIClient, *kafkamgmtclient.KafkaRequest, error)
	HttpClient() *http.Client
}
