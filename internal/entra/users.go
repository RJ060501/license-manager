package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// User represents the subset of the Microsoft Graph user object that the
// license manager needs.
type User struct {
	ID                string            `json:"id"`
	DisplayName       string            `json:"displayName"`
	UserPrincipalName string            `json:"userPrincipalName"`
	AccountEnabled    bool              `json:"accountEnabled"`
	AssignedLicenses  []AssignedLicense `json:"assignedLicenses"`
}

// AssignedLicense represents a Microsoft SKU assigned to a user.
type AssignedLicense struct {
	SKUID string `json:"skuId"`
}

// GetUsers retrieves all Entra users and follows Microsoft Graph pagination.
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	url := c.baseURL +
		"/users?$select=id,displayName,userPrincipalName,accountEnabled,assignedLicenses"

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
				"failed to create Graph users request: %w",
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
				"Microsoft Graph returned status %s while retrieving users",
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
				"failed to decode Graph users response: %w",
				err,
			)
		}

		users = append(users, result.Value...)

		url = result.NextLink
	}

	return users, nil
}
