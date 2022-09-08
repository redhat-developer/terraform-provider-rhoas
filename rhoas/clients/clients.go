package clients

import (
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	serviceAccounts "github.com/redhat-developer/app-services-sdk-go/serviceaccountmgmt/apiv1/client"
)

type Clients struct {
	KafkaClient          *kafkamgmtclient.APIClient
	ServiceAccountClient *serviceAccounts.APIClient
}
