package openfga

import (
	"fmt"

	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	fgaclient "github.com/openfga/go-sdk/client"
	fgacredentials "github.com/openfga/go-sdk/credentials"
)

type Client struct {
	sdk *fgaclient.OpenFgaClient
}

func NewClient(cfg *conf.App_OpenFGA) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("openfga: nil config")
	}
	if cfg.ApiUrl == "" || cfg.StoreId == "" {
		return nil, fmt.Errorf("openfga: api_url and store_id are required")
	}

	cc := &fgaclient.ClientConfiguration{
		ApiUrl:               cfg.ApiUrl,
		StoreId:              cfg.StoreId,
		AuthorizationModelId: cfg.ModelId,
	}
	if cfg.ApiToken != "" {
		cc.Credentials = &fgacredentials.Credentials{
			Method: fgacredentials.CredentialsMethodApiToken,
			Config: &fgacredentials.Config{ApiToken: cfg.ApiToken},
		}
	}

	sdk, err := fgaclient.NewSdkClient(cc)
	if err != nil {
		return nil, fmt.Errorf("openfga: %w", err)
	}
	return &Client{sdk: sdk}, nil
}
