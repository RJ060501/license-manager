package entra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/RJ060501/license-manager/internal/config"
)

// Client contains everything required to communicate with Microsoft Graph.
type Client struct {
	credential *azidentity.ClientSecretCredential
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a Microsoft Graph client using the application's
// registered Entra credentials.
func NewClient(cfg *config.Config) (*Client, error) {
	credential, err := azidentity.NewClientSecretCredential(
		cfg.TenantID,
		cfg.ClientID,
		cfg.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Entra credential: %w",
			err,
		)
	}

	return &Client{
		credential: credential,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://graph.microsoft.com/v1.0",
	}, nil
}

// getAccessToken obtains an app-only Microsoft Graph bearer token.
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	token, err := c.credential.GetToken(
		ctx,
		policy.TokenRequestOptions{
			Scopes: []string{
				"https://graph.microsoft.com/.default",
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to acquire Microsoft Graph token: %w",
			err,
		)
	}

	return token.Token, nil
}
