> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Social Features

> Access user profiles, view leaderboards, and track the global trade feed on Jupiter Prediction Market.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc covers the social features of Jupiter Prediction Market: user profiles, leaderboards, and the global trade feed.

## Prerequisite

To use social features, you need:

* An API key from the [Portal](https://developers.jup.ag/portal)

> [!TIP]
>
> **API REFERENCE**
>
> For complete endpoint specifications, see the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction).

***

## User Profiles

Use `GET /profiles/{ownerPubkey}` to retrieve a user's prediction market statistics and performance metrics.

```js theme={null}
const traderPubkey = 'TRADER_WALLET_PUBLIC_KEY';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/profiles/${traderPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const profile = await response.json();
console.log(profile);
```

Example response:

```json expandable theme={null}
{
  "ownerPubkey": "TRADER_WALLET_PUBLIC_KEY",
  "realizedPnlUsd": "15000000",
  "totalVolumeUsd": "250000000",
  "predictionsCount": "45",
  "correctPredictions": "28",
  "wrongPredictions": "17",
  "totalActiveContracts": "150",
  "totalPositionsValueUsd": "85000000"
}
```

| Field | Description |
| :- | :- |
| `realizedPnlUsd` | Total realized profit/loss in micro USD |
| `totalVolumeUsd` | Total trading volume in micro USD |
| `predictionsCount` | Total number of predictions made |
| `correctPredictions` | Number of winning predictions |
| `wrongPredictions` | Number of losing predictions |
| `totalActiveContracts` | Current open contract count |
| `totalPositionsValueUsd` | Current portfolio value in micro USD |

***

## P\&L History

Use `GET /profiles/{ownerPubkey}/pnl-history` to retrieve time-series P\&L data for charting.

| Parameter | Type | Description |
| :- | :- | :- |
| `interval` | string | Time interval: `24h`, `1w`, or `1m` |
| `count` | number | Number of data points (default: 10) |

```js theme={null}
const traderPubkey = 'TRADER_WALLET_PUBLIC_KEY';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/profiles/${traderPubkey}/pnl-history?` +
  new URLSearchParams({
    interval: '1w',
    count: '12' // 12 weeks of data
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const pnlHistory = await response.json();
```

Example response:

```json expandable theme={null}
{
  "ownerPubkey": "TRADER_WALLET_PUBLIC_KEY",
  "history": [
    { "timestamp": 1704067200, "realizedPnlUsd": "5000000" },
    { "timestamp": 1704672000, "realizedPnlUsd": "7500000" },
    { "timestamp": 1705276800, "realizedPnlUsd": "6800000" },
    { "timestamp": 1705881600, "realizedPnlUsd": "12000000" },
    { "timestamp": 1706486400, "realizedPnlUsd": "15000000" }
  ]
}
```

***

## Trade Feed

Use `GET /trades` to retrieve recent trades across the platform. This provides a global activity feed of filled orders.

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/trades',
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const trades = await response.json();
```

Example response:

```json expandable theme={null}
{
  "data": [
    {
      "id": 425993,
      "ownerPubkey": "trader-pubkey-1",
      "marketId": "market-456",
      "message": "trader-pubkey-1 bought Yes for SOL $500 by Q2 2026 at $0.72 ($18.00)",
      "timestamp": 1704067260,
      "action": "buy",
      "side": "yes",
      "eventTitle": "Will SOL reach $500?",
      "marketTitle": "SOL $500 by Q2 2026",
      "amountUsd": "18000000",
      "priceUsd": "720000",
      "eventImageUrl": "https://...",
      "eventId": "event-123"
    },
    {
      "id": 425992,
      "ownerPubkey": "trader-pubkey-2",
      "marketId": "market-012",
      "message": "trader-pubkey-2 sold No for BTC halving before April 2024 at $0.45 ($4.50)",
      "timestamp": 1704067200,
      "action": "sell",
      "side": "no",
      "eventTitle": "Bitcoin halving date",
      "marketTitle": "BTC halving before April 2024",
      "amountUsd": "4500000",
      "priceUsd": "450000",
      "eventImageUrl": "https://...",
      "eventId": "event-789"
    }
  ]
}
```

| Field | Description |
| :- | :- |
| `id` | Unique trade identifier |
| `ownerPubkey` | Trader's wallet address |
| `marketId` | Market identifier |
| `message` | Human-readable trade summary |
| `timestamp` | Unix timestamp (seconds) of the trade |
| `action` | `buy` or `sell` |
| `side` | `yes` or `no` |
| `priceUsd` | Execution price in micro USD |
| `amountUsd` | Total trade value in micro USD |
| `eventTitle` | Parent event title |
| `marketTitle` | Market title |
| `eventImageUrl` | Image URL for the parent event |
| `eventId` | Parent event identifier |

***

## Leaderboards

Use `GET /leaderboards` to retrieve competitive rankings.

| Parameter | Type | Description |
| :- | :- | :- |
| `period` | string | Time period: `all_time`, `weekly`, `monthly` |
| `metric` | string | Ranking metric: `pnl`, `volume`, `win_rate` |
| `limit` | number | Number of results (1-100) |

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/leaderboards?' + new URLSearchParams({
    period: 'weekly',
    metric: 'pnl',
    limit: '10'
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const leaderboard = await response.json();
```

Example response:

```json expandable theme={null}
{
  "data": [
    {
      "ownerPubkey": "top-trader-pubkey",
      "realizedPnlUsd": "50000000",
      "totalVolumeUsd": "500000000",
      "predictionsCount": 40,
      "correctPredictions": 30,
      "wrongPredictions": 10,
      "winRatePct": "75.00",
      "period": "weekly",
      "periodStart": "2026-02-24T00:00:00.000Z",
      "periodEnd": "2026-03-03T00:00:00.000Z"
    },
    {
      "ownerPubkey": "second-trader-pubkey",
      "realizedPnlUsd": "35000000",
      "totalVolumeUsd": "400000000",
      "predictionsCount": 55,
      "correctPredictions": 37,
      "wrongPredictions": 18,
      "winRatePct": "67.27",
      "period": "weekly",
      "periodStart": "2026-02-24T00:00:00.000Z",
      "periodEnd": "2026-03-03T00:00:00.000Z"
    }
  ],
  "summary": {
    "all_time": {
      "totalVolumeUsd": "7009708479895",
      "predictionsCount": 31202
    },
    "weekly": {
      "totalVolumeUsd": "186403707003",
      "predictionsCount": 971
    },
    "monthly": {
      "totalVolumeUsd": "269887658105",
      "predictionsCount": 1743
    }
  }
}
```

| Metric | Description |
| :- | :- |
| `pnl` | Ranked by realized profit/loss |
| `volume` | Ranked by total trading volume |
| `win_rate` | Ranked by prediction accuracy |

***

## Use Cases

### Analytics Dashboard

Combine leaderboards with profiles for competitive analysis:

```js theme={null}
// Get top performers
const topTraders = await fetch(
  'https://api.jup.ag/prediction/v1/leaderboards?' + new URLSearchParams({
    period: 'monthly',
    metric: 'pnl',
    limit: '5'
  }),
  { headers: { 'x-api-key': 'your-api-key' } }
).then(r => r.json());

// Get detailed P&L history for top trader
const topTraderHistory = await fetch(
  `https://api.jup.ag/prediction/v1/profiles/${topTraders.data[0].ownerPubkey}/pnl-history?` +
  new URLSearchParams({ interval: '1w', count: '4' }),
  { headers: { 'x-api-key': 'your-api-key' } }
).then(r => r.json());

console.log('Top trader weekly P&L trend:', topTraderHistory.history);
```

## Summary

| Feature | Endpoint | Description |
| :- | :- | :- |
| Profiles | `GET /profiles/{pubkey}` | User stats and performance |
| P\&L History | `GET /profiles/{pubkey}/pnl-history` | Time-series P\&L data |
| Trade Feed | `GET /trades` | Global recent trades |
| Leaderboards | `GET /leaderboards` | Competitive rankings |


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.