package main

import (
	"context"
	"fmt"
	"log"

	"github.com/RJ060501/license-manager/internal/config"
	"github.com/RJ060501/license-manager/internal/entra"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf(
			"failed to load configuration: %v",
			err,
		)
	}

	entraClient, err := entra.NewClient(cfg)
	if err != nil {
		log.Fatalf(
			"failed to create Entra client: %v",
			err,
		)
	}

	ctx := context.Background()

	users, err := entraClient.GetUsers(ctx)
	if err != nil {
		log.Fatalf(
			"failed to retrieve users: %v",
			err,
		)
	}

	fmt.Printf(
		"Retrieved %d users.\n\n",
		len(users),
	)

	limit := 10

	if len(users) < limit {
		limit = len(users)
	}

	for _, user := range users[:limit] {
		fmt.Printf(
			"%s | %s | Enabled: %t\n",
			user.DisplayName,
			user.UserPrincipalName,
			user.AccountEnabled,
		)
	}

	skus, err := entraClient.GetSubscribedSKUs(ctx)
	if err != nil {
		log.Fatalf(
			"failed to retrieve subscribed SKUs: %v",
			err,
		)
	}

	fmt.Printf(
		"\nRetrieved %d subscribed SKUs.\n\n",
		len(skus),
	)

	for _, sku := range skus {
		fmt.Printf(
			"%s | %s | Consumed: %d\n",
			sku.PartNumber,
			sku.SKUID,
			sku.ConsumedUnits,
		)
	}
}
