package entra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SubscribedSKU represents a Microsoft product/license SKU that is available
// in the organization's Microsoft 365 tenant.
//
// Microsoft Graph returns license assignments using SKU IDs (GUIDs). The
// /subscribedSkus endpoint allows us to translate those GUIDs into readable
// Microsoft SKU part numbers.
//
// Example:
//
//	SKUID:         "05e9a617-0261-4cee-bb44-138d3ef5d965"
//	PartNumber:    "SPE_E3"
//	ConsumedUnits: 175
//
// Later, we will use this information to map the SKU IDs assigned to each
// Entra user to recognizable products such as Microsoft 365 E3, Visio,
// Project, Teams Phone, etc.
type SubscribedSKU struct {
	SKUID         string `json:"skuId"`
	PartNumber    string `json:"skuPartNumber"`
	ConsumedUnits int    `json:"consumedUnits"`
}

// GetSubscribedSKUs retrieves all Microsoft license SKUs subscribed to by
// the tenant.
//
// The returned list represents the Microsoft products available in the
// organization's tenant. This does not tell us which user has which license;
// individual assignments are returned through the assignedLicenses property
// on the Microsoft Graph user object.
//
// We will later correlate:
//
//	User.AssignedLicenses[].SKUID
//
// with:
//
//	SubscribedSKU.SKUID
//
// to determine the readable license name for each user.
func (c *Client) GetSubscribedSKUs(
	ctx context.Context,
) ([]SubscribedSKU, error) {
	// Acquire an app-only bearer token for Microsoft Graph.
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Microsoft Graph endpoint containing the tenant's subscribed license SKUs.
	url := c.baseURL + "/subscribedSkus"

	// Create the GET request and attach the caller's context.
	//
	// Using the context allows cancellation or deadlines from main.go (or
	// another caller) to propagate into the HTTP request.
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create subscribed SKUs request: %w",
			err,
		)
	}

	// Microsoft Graph requires the access token in the Authorization header.
	request.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	// Tell Graph that we expect a JSON response.
	request.Header.Set(
		"Accept",
		"application/json",
	)

	// Send the request using the shared HTTP client configured in client.go.
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve subscribed SKUs: %w",
			err,
		)
	}

	// Make sure the response body is closed when this function finishes.
	//
	// This is especially important for allowing Go's HTTP transport to reuse
	// TCP connections efficiently.
	defer response.Body.Close()

	// A successful network request does not necessarily mean Microsoft Graph
	// accepted the request. Graph may return statuses such as:
	//
	//	401 Unauthorized
	//	403 Forbidden
	//	429 Too Many Requests
	//
	// For now, treat anything outside the 2xx range as an error.
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"Microsoft Graph returned status %s while retrieving subscribed SKUs",
			response.Status,
		)
	}

	// Microsoft Graph collection responses generally look like:
	//
	// {
	//     "value": [
	//         {...},
	//         {...}
	//     ]
	// }
	//
	// We only need the "value" property for this endpoint.
	var result struct {
		Value []SubscribedSKU `json:"value"`
	}

	// Decode Microsoft's JSON response directly into our Go structs.
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"failed to decode subscribed SKUs response: %w",
			err,
		)
	}

	return result.Value, nil
}

// BuildSKULookup creates a map that translates Microsoft SKU IDs into
// readable SKU part numbers.
//
// Example:
//
//	"05e9a617-..." -> "SPE_E3"
//
// The SKU ID is what Microsoft stores on each user's assignedLicenses array.
// The SKU part number is much easier for humans to understand.
func BuildSKULookup(skus []SubscribedSKU) map[string]string {
	// Create an empty map where:
	//
	//	key   = Microsoft SKU GUID
	//	value = Microsoft SKU part number
	lookup := make(map[string]string)

	// Loop through every SKU returned by Microsoft Graph.
	for _, sku := range skus {
		lookup[sku.SKUID] = sku.PartNumber
	}

	return lookup
}
