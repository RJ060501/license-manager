package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/RJ060501/license-manager/internal/config"
)

type Client struct {
	credential *azidentity.ClientSecretCredential
	httpClient *http.Client
	baseURL    string
}

func NewClient(cfg *config.Config) (*Client, error) {
	credential, err := azidentity.NewClientSecretCredential(
		cfg.TenantID,
		cfg.ClientID,
		cfg.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Entra credential: %w", err)
	}

	return &Client{
		credential: credential,
		httpClient: &http.Client{},
		baseURL:    "https://graph.microsoft.com/v1.0",
	}, nil
}

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

type User struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	UserPrincipalName string `json:"userPrincipalName"`
	AccountEnabled    bool   `json:"accountEnabled"`
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := c.baseURL +
		"/users?$select=id,displayName,userPrincipalName,accountEnabled"

	var users []User

	for url != "" {
		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			url,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create Graph request: %w",
				err,
			)
		}

		request.Header.Set(
			"Authorization",
			"Bearer "+token,
		)

		request.Header.Set(
			"Accept",
			"application/json",
		)

		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve Entra users: %w",
				err,
			)
		}

		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response.Body.Close()

			return nil, fmt.Errorf(
				"Microsoft Graph returned status %s",
				response.Status,
			)
		}

		var result struct {
			Value    []User `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		err = json.NewDecoder(response.Body).Decode(&result)

		response.Body.Close()

		if err != nil {
			return nil, fmt.Errorf(
				"failed to decode Graph response: %w",
				err,
			)
		}

		users = append(users, result.Value...)

		url = result.NextLink
	}

	return users, nil
}
