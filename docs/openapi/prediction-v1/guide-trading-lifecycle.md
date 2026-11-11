> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Trading Lifecycle

> The end-to-end sequence of Prediction API calls for a trade, and how data flows between them.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This page maps a trade from start to finish: which Prediction API endpoints to call, in what order, and what data flows from one call into the next. Each endpoint also has its own detail page, linked at every step.

For a hands-on build with a full runnable app, video walkthrough, and demo source, see [How to Build a Prediction Market App on Solana](https://developers.jup.ag/docs/guides/how-to-build-a-prediction-market-app-on-solana.md).

## Prerequisite

* An API key from the [Portal](https://developers.jup.ag/portal). All requests require the `x-api-key` header.
* A Solana wallet with funds (JupUSD or USDC) for the trading steps.

## Lifecycle at a Glance

A trade moves through discovery, a pre-trade check, the order itself, then tracking and settlement. The key to integrating is the data hand-off: each call produces the identifier the next call needs.

```text theme={null}
GET /events/search?query=…        ─▶ event.markets[].marketId
GET /markets/{marketId}           ─▶ current YES/NO prices
GET /orderbook/{marketId}         ─▶ available depth
GET /trading-status               ─▶ exchange open?
POST /orders                      ─▶ order.orderPubkey + unsigned tx ─▶ sign & submit
GET /orders/status/{orderPubkey}  ─▶ fill status
GET /positions                    ─▶ open position + P&L
POST /positions/{positionPubkey}/claim ─▶ payout (after resolution)
```

## Step by Step

### 1. Discover events

Start with `GET /events` (filter by `category`, `filter`, `tags`) or `GET /events/search?query=` for keyword search. Each event in the response carries a `markets[]` array.

See [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md) for parameters and the response shape.

### 2. Pick a market

Read the `marketId` you want from `event.markets[]`. An event holds one market per outcome (for example, one market per team), so select the specific outcome the user is trading.

### 3. Read current pricing

Call `GET /markets/{marketId}` to get fresh `buyYesPriceUsd`, `buyNoPriceUsd`, `sellYesPriceUsd`, and `sellNoPriceUsd`. Prices are in micro USD (`1000000` = \$1.00) and reflect implied probability.

> [!TIP]
>
> Prices move. Re-fetch `GET /markets/{marketId}` immediately before placing an order rather than relying on pricing embedded in an earlier list response.

### 4. Check orderbook depth

Call `GET /orderbook/{marketId}` to see bid/ask depth on each side. Depth tells you how much size can fill near the quoted price, which matters for larger orders.

See [Orderbook](https://developers.jup.ag/docs/prediction/events-and-markets.md#get-orderbook-data) for the array format.

### 5. Confirm the exchange is open

Call `GET /trading-status` and confirm `trading_active` is `true` before attempting an order. Also confirm the market itself is tradeable (`market.status === 'open'`).

### 6. Place the order

Call `POST /orders` with the `marketId`, side (`isYes`), `isBuy: true`, and a deposit. It returns an unsigned transaction. Deserialize it, have the user sign, and submit it to Solana.

Full request parameters, the \$5 minimum, and the sign-and-submit code are on [Open Positions](https://developers.jup.ag/docs/prediction/open-positions.md).

### 7. Track the fill

After submitting, poll `GET /orders/status/{orderPubkey}` for the fill status (`created`, `partiallyfilled`, `filled`, `failed`) and read the resulting holding from `GET /positions`.

See [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md) and [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md).

### 8. Settle or claim

When the market resolves, a winning position becomes claimable. Call `POST /positions/{positionPubkey}/claim` to build the payout transaction, then sign and submit it. Some markets settle automatically and need no claim.

See [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md).

## Related

> [card] [Build a Prediction Market App](https://developers.jup.ag/docs/guides/how-to-build-a-prediction-market-app-on-solana.md)
>
> Full hands-on guide with runnable code, a video walkthrough, and demo app source.

> [card] [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md)
>
> Discover events, read pricing, and inspect orderbook depth.

> [card] [Open Positions](https://developers.jup.ag/docs/prediction/open-positions.md)
>
> Create, sign, and submit a buy order.

> [card] [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md)
>
> Sell contracts and close positions.

> [card] [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md)
>
> Claim winnings after a market resolves.

> [card] [API Reference](https://developers.jup.ag/docs/api-reference/prediction)
>
> Full endpoint specifications.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.