> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Events & Markets

> Discover prediction market events, search by keywords, and get market details with pricing information.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc covers how to discover prediction market events, search for specific topics, and retrieve detailed market information including pricing data.

For conceptual background on how events, markets, and pricing work, see the [Prediction Market Overview](https://developers.jup.ag/docs/prediction.md).

## Prerequisite

To query events and markets data, you need:

* An API key from the [Portal](https://developers.jup.ag/portal)

> [!TIP]
>
> **API REFERENCE**
>
> For complete endpoint specifications, see the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction).

***

## List Events

Use `GET /events` to retrieve a paginated list of prediction market events. You can filter by category, provider, and status.

| Parameter | Type | Description |
| :- | :- | :- |
| `category` | string | Filter by category: `crypto`, `sports`, `politics`, `esports`, `culture`, `economics`, `tech` |
| `subcategory` | string | Filter by subcategory within the category |
| `tags` | string | Filter by tag (e.g. `soccer`) |
| `filter` | string | Filter by status: `new`, `live`, `trending`, `upcoming` |
| `sortBy` | string | Sort field |
| `sortDirection` | string | Sort direction: `asc` or `desc` |
| `includeMarkets` | boolean | Include nested markets in response |
| `includeAllMarkets` | boolean | Include all sports market types (moneyline, spread, totals) and Saba markets. Defaults to `false` (moneyline only); F1 events always include every type |
| `start` | number | Pagination start index |
| `end` | number | Pagination end index |

Example: List Trending Crypto Events

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/events?' + new URLSearchParams({
    category: 'crypto',
    filter: 'trending',
    includeMarkets: 'true'
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const events = await response.json();
console.log(events);
```

***

## Search Events

Use `GET /events/search` to find events by keyword.

| Parameter | Type | Description |
| :- | :- | :- |
| `query` | string | Search term |
| `limit` | number | Maximum results to return |

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/events/search?' + new URLSearchParams({
    query: 'nba',
    limit: '10'
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const searchResults = await response.json();
```

***

## Get Event Details

Use `GET /events/{eventId}` to retrieve full details for a specific event, including all its markets.

```js theme={null}
const eventId = 'event-123';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/events/${eventId}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const event = await response.json();
```

***

## Get Suggested Events

Use `GET /events/suggested/{pubkey}` to get personalized event recommendations based on a user's trading history.

```js theme={null}
const userPubkey = 'YOUR_WALLET_PUBLIC_KEY';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/events/suggested/${userPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const suggestedEvents = await response.json();
```

***

## Get Market Details

Use `GET /markets/{marketId}` to retrieve detailed information about a specific market, including current pricing.

> [!NOTE]
>
> There can be multiple markets for the same event.

```js theme={null}
const marketId = 'market-456';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/markets/${marketId}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const market = await response.json();
console.log(market);
```

Example response:

```json expandable theme={null}
{
  "marketId": "market-456",
  "eventId": "event-123",
  "provider": "polymarket",
  "title": "SOL $500 by Q2 2026",
  "status": "open",
  "result": null,
  "openTime": 1704067200,
  "closeTime": 1719792000,
  "resolveAt": null,
  "marketResultPubkey": null,
  "imageUrl": null,
  "rulesPrimary": "Resolves YES if SOL reaches $500...",
  "rulesSecondary": "",
  "outcomes": ["Yes", "No"],
  "isTeamMarket": false,
  "pricing": {
    "buyYesPriceUsd": 650000,
    "buyNoPriceUsd": 380000,
    "sellYesPriceUsd": 620000,
    "sellNoPriceUsd": 350000,
    "volume": 450000
  }
}
```

Market fields are returned flat. There is no nested `metadata` object. `resolveAt` stays `null` while the market is open and becomes an ISO 8601 string once it resolves.

### Understanding Prices

> [!NOTE]
>
> **Prices are in JupUSD or USDC native token units**
>
> All prices in the API are denominated in native token units where **1,000,000 native token units for JupUSD or USDC = \$1.00**.
>
> For example:
>
> * `650000` = \$0.65 (650000 / 1,000,000 = \$0.65)
> * `380000` = \$0.38 (380000 / 1,000,000 = \$0.38)
> * `10000` = \$0.01 (10000 / 1,000,000 = \$0.01)

The market response includes four price fields:

| Field | Description |
| :- | :- |
| `buyYesPriceUsd` | Cost to buy a YES contract |
| `sellYesPriceUsd` | Amount received when selling a YES contract |
| `buyNoPriceUsd` | Cost to buy a NO contract |
| `sellNoPriceUsd` | Amount received when selling a NO contract |
| `volume` | Trading volume for this market |

> [!TIP]
>
> **Price as Probability**
>
> Market prices reflect the implied probability of an outcome. A `buyYesPriceUsd` of `650000` (\$0.65) suggests the market believes there's roughly a 65% chance of YES being the outcome.

### Market Status

| Status | Description |
| :- | :- |
| `open` | Market is actively trading |
| `closed` | Trading has stopped, awaiting settlement |
| `cancelled` | Market voided, positions returned |

### Market Result

| Result | Description |
| :- | :- |
| `null` | Market not yet resolved |
| `yes` | YES outcome confirmed |
| `no` | NO outcome confirmed |

***

## Get Orderbook Data

Use `GET /orderbook/{marketId}` to retrieve the current orderbook with bid/ask depth.

The response has four arrays. Each entry is a `[price, quantity]` pair:

| Field | Description |
| :- | :- |
| `yes` | YES bids: `[price_cents, quantity]`. Price is in cents (e.g. `1` = \$0.01). `price_cents` is rounded, so sub-cent levels show `0`; use `yes_dollars` for the exact price. |
| `no` | NO bids: `[price_cents, quantity]`. Same format as `yes`. |
| `yes_dollars` | YES bids with price as decimal string: `["0.0100", quantity]`. |
| `no_dollars` | NO bids with price as decimal string. Same format as `yes_dollars`. |

Arrays are sorted by price ascending. Quantities are in base units and may be fractional.

Example: Get orderbook data for a specific market

```js theme={null}
const marketId = 'market-456';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/orderbook/${marketId}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const orderbook = await response.json();
```

**Example response:**

```json expandable theme={null}
{
  "yes": [
    [1, 346014],
    [2, 10568],
    [3, 12300]
  ],
  "no": [
    [1, 219],
    [2, 14],
    [3, 4400]
  ],
  "yes_dollars": [
    ["0.0100", 346014],
    ["0.0200", 10568],
    ["0.0300", 12300]
  ],
  "no_dollars": [
    ["0.0100", 219],
    ["0.0200", 14],
    ["0.0300", 4400]
  ]
}
```

***

## Get Live Scores

For sports events, use `GET /events/scores` to fetch live scores for one or more events, or `GET /events/{eventId}/score` for a single event. Events without a score row are omitted from the list response.

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/events/scores?' + new URLSearchParams({
    eventIds: 'POLY-351731,POLY-30615' // comma-separated, max 100
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const { data } = await response.json();
```

Example response:

```json expandable theme={null}
{
  "data": [
    {
      "eventId": "POLY-351731",
      "gameId": "90086923",
      "leagueAbbreviation": "fifwc",
      "homeTeam": "France",
      "awayTeam": "Senegal",
      "score": "2-0",
      "period": "2H",
      "elapsed": "90",
      "status": "InProgress",
      "live": true,
      "ended": false,
      "finishedTimestamp": null,
      "updatedAt": "2026-06-16T20:53:39.834Z"
    }
  ]
}
```

| Field | Description |
| :- | :- |
| `score` | Raw score string from the provider (e.g. `"2-0"`) |
| `period` | Current period (e.g. `"2H"`) |
| `elapsed` | Elapsed time within the game (e.g. `"90"`) |
| `status` | Provider status string (e.g. `"InProgress"`) |
| `live` | `true` while the game is in progress |
| `ended` | `true` once the game has finished |
| `finishedTimestamp` | ISO 8601 timestamp when the game finished, if ended |
| `updatedAt` | ISO 8601 timestamp of the last score update |

***

## Check Trading Status

Use `GET /trading-status` to verify if the exchange is actively trading. This is useful for conditional logic before placing orders.

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/trading-status',
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const status = await response.json();
console.log('Trading active:', status.trading_active);
```

***

## What's Next

Now that you know how to discover events and markets, you're ready to open positions.

> [card] [Open Positions](https://developers.jup.ag/docs/prediction/open-positions.md)
>
> Create buy orders for YES or NO contracts


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.