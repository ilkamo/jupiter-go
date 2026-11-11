// Command prediction checks whether the Prediction Market exchange is open and
// prints trending events with their YES prices. It reads only; nothing is signed.
//
// The API is beta upstream and blocks United States and South Korea IP addresses.
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/prediction"
)

const (
	operationTimeout = 15 * time.Second
	trendingEvents   = 5

	// microUSDPerDollar converts micro USD prices: 850000 is $0.85.
	microUSDPerDollar = 1_000_000
	priceDecimals     = 2
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	status, err := client.Prediction.TradingStatus(ctx)
	if err != nil {
		return fmt.Errorf("trading status: %w", err)
	}

	if _, err := fmt.Printf("Trading active: %t\n", status.TradingActive); err != nil {
		return err
	}

	end := uint32(trendingEvents)
	events, err := client.Prediction.Events(ctx, prediction.EventsRequest{
		Filter:         prediction.FilterTrending,
		IncludeMarkets: true,
		Page:           prediction.Page{End: &end},
	})
	if err != nil {
		return fmt.Errorf("events: %w", err)
	}

	for _, event := range events.Events {
		if _, err := fmt.Printf("%s (%s)\n", event.EventMetadata.Title, event.EventID); err != nil {
			return err
		}

		for _, market := range event.Markets {
			if err := printMarket(market); err != nil {
				return err
			}
		}
	}

	return nil
}

func printMarket(market prediction.Market) error {
	if market.Pricing == nil || market.Pricing.BuyYesPriceUSD == "" {
		_, err := fmt.Printf("  %s: no price\n", market.Title)

		return err
	}

	price, ok := new(big.Rat).SetString(market.Pricing.BuyYesPriceUSD.String())
	if !ok {
		return fmt.Errorf("market %s: invalid price %q", market.MarketID, market.Pricing.BuyYesPriceUSD)
	}

	dollars := price.Quo(price, big.NewRat(microUSDPerDollar, 1))
	_, err := fmt.Printf(
		"  %s: YES at $%s (%s)\n",
		market.Title,
		dollars.FloatString(priceDecimals),
		market.Status,
	)

	return err
}
