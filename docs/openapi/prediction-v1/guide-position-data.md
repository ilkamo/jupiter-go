> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Position Data & History

> Query positions, orders, and transaction history to track your prediction market activity and P&L.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc covers the data endpoints for querying positions, orders, and transaction history. Use these endpoints to build dashboards, track P\&L, and maintain audit trails.

## Prerequisite

To query position data, you need:

* An API key from the [Portal](https://developers.jup.ag/portal)

> [!TIP]
>
> **API REFERENCE**
>
> For complete endpoint specifications, see the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction).

***

## Querying Positions

Use `GET /positions` to retrieve position data with comprehensive P\&L information.

| Parameter | Type | Description |
| :- | :- | :- |
| `ownerPubkey` | string | Wallet public key (required) |
| `marketPubkey` | string | Filter by market account pubkey |
| `marketId` | string | Filter by market ID |
| `isYes` | boolean | Filter by position side |
| `start` | number | Pagination start index |
| `end` | number | Pagination end index |

List endpoints (`/positions`, `/orders`, `/history`) return a `pagination` object alongside `data`:

| Field | Type | Description |
| :- | :- | :- |
| `start` | integer | Start index of the current page |
| `end` | integer | End index of the current page |
| `total` | integer | Total number of results available |
| `hasNext` | boolean | `true` if more results exist beyond the current page |

### Position Fields

Each position includes identity, state, and P\&L fields. All USD values are in micro USD (1,000,000 = \$1.00).

#### Identity and State

| Field | Type | Description |
| :- | :- | :- |
| `pubkey` | string | Position account public key |
| `ownerPubkey` | string | Position owner wallet public key |
| `marketId` | string | External market identifier |
| `isYes` | boolean | `true` if this is a YES position, `false` for NO |
| `contracts` | string | Legacy whole-contract quantity. Contracts are fractional: use `contractsMicro` (`1000000` = 1) or `contractsDecimal` for the exact amount |
| `contractsMicro` | string | Contracts held in micro-contract units |
| `contractsDecimal` | string | Contracts held as a decimal string |
| `openOrders` | integer | Number of currently open orders for this position |
| `claimable` | boolean | `true` when the position qualifies for payout claim |
| `claimed` | boolean | Whether payout has already been claimed |
| `claimedUsd` | string | Amount claimed in micro USD |
| `openedAt` | integer | Unix timestamp (seconds) when the position was opened |
| `updatedAt` | integer | Unix timestamp (seconds) of the last position update |
| `claimableAt` | integer, nullable | Unix timestamp (seconds) when position becomes claimable |
| `settlementDate` | integer, nullable | Unix timestamp (seconds) when the market settles |

#### Pricing and P\&L

| Field | Type | Description |
| :- | :- | :- |
| `totalCostUsd` | string | Total cost basis in micro USD |
| `sizeUsd` | string | Alias of `totalCostUsd` in micro USD |
| `avgPriceUsd` | string | Average entry price per contract in micro USD |
| `markPriceUsd` | string, nullable | Current mark price per contract in micro USD; `null` when market is closed |
| `sellPriceUsd` | string, nullable | Current best exit price for this side in micro USD; `null` when unavailable |
| `valueUsd` | string, nullable | Current mark-to-market value in micro USD; `null` when market is closed |
| `payoutUsd` | string | Notional payout if position wins (contracts x \$1) |
| `pnlUsd` | string, nullable | Unrealized P\&L in micro USD; `null` when market is closed |
| `pnlUsdPercent` | number, nullable | Unrealized P\&L as percentage of cost basis |
| `pnlUsdAfterFees` | string, nullable | Unrealized P\&L after estimated fees in micro USD |
| `pnlUsdAfterFeesPercent` | number, nullable | Unrealized P\&L after fees as percentage of cost basis |
| `realizedPnlUsd` | number | Realized P\&L from closed portions in micro USD |
| `feesPaidUsd` | string | Total fees paid in micro USD |

#### Metadata

| Field | Type | Description |
| :- | :- | :- |
| `eventMetadata` | object | Parent event metadata: `eventId`, `title`, `subtitle`, `imageUrl`, `isLive`, `isActive`, `category`, `subcategory`, `beginAt`, `closeCondition` |
| `marketMetadata` | object | Market metadata: `marketId`, `eventId`, `title`, `subtitle`, `description`, `status`, `result`, `openTime`, `closeTime` |

### Get All Positions

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/positions?' + new URLSearchParams({
    ownerPubkey: wallet.publicKey.toString()
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const positions = await response.json();
```

> [accordion] Aggregating Portfolio P\&L
>
> ```js theme={null}
> const positions = await fetch(
>   'https://api.jup.ag/prediction/v1/positions?' + new URLSearchParams({
>     ownerPubkey: wallet.publicKey.toString()
>   }),
>   {
>     headers: { 'x-api-key': 'your-api-key' }
>   }
> ).then(r => r.json());
>
> // Calculate totals
> let totalValue = 0;
> let totalCost = 0;
> let totalPnl = 0;
>
> for (const position of positions.data) {
>   totalValue += parseInt(position.valueUsd) || 0;
>   totalCost += parseInt(position.totalCostUsd) || 0;
>   totalPnl += parseInt(position.pnlUsd) || 0;
> }
>
> console.log('Portfolio Summary:');
> console.log(`Total Value: $${(totalValue / 1_000_000).toFixed(2)}`);
> console.log(`Total Cost: $${(totalCost / 1_000_000).toFixed(2)}`);
> console.log(`Total P&L: $${(totalPnl / 1_000_000).toFixed(2)}`);
> ```

### Get a Specific Position

```js theme={null}
const positionPubkey = 'position-pubkey-789';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/positions/${positionPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const position = await response.json();
```

***

## Querying Orders

Use `GET /orders` to retrieve order history and fill details.

| Parameter | Type | Description |
| :- | :- | :- |
| `ownerPubkey` | string | Wallet public key |
| `start` | number | Pagination start (or timestamp) |
| `end` | number | Pagination end (or timestamp) |

### Order Fill Details

| Field | Description |
| :- | :- |
| `status` | `pending`, `filled`, or `failed` |
| `filledContracts` | Number of contracts executed |
| `avgFillPriceUsd` | Average execution price |
| `totalCostUsd` | Total cost of filled portion |
| `feesPaidUsd` | Fees charged |
| `createdAt` | Unix timestamp of order creation |
| `filledAt` | Unix timestamp of fill completion |

### Get All Orders

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/orders?' + new URLSearchParams({
    ownerPubkey: wallet.publicKey.toString()
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const orders = await response.json();
```

### Get Order Status

Use `GET /orders/status/{orderPubkey}` for detailed status and event history.

```js theme={null}
const orderPubkey = 'order-pubkey-123';
const response = await fetch(
  `https://api.jup.ag/prediction/v1/orders/status/${orderPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const status = await response.json();
```

***

## Transaction History

Use `GET /history` to retrieve a complete audit trail of all prediction market activity.

| Parameter | Type | Description |
| :- | :- | :- |
| `ownerPubkey` | string | Wallet public key |
| `positionPubkey` | string | Filter by specific position |
| `start` | number | Start timestamp or index |
| `end` | number | End timestamp or index |
| `id` | string | Filter by specific event ID |

### Event Types

| Event Type | Description |
| :- | :- |
| `order_created` | New order placed on-chain |
| `order_filled` | Order matched and executed |
| `order_failed` | Order could not be filled |
| `order_closed` | Order account closed after completion |
| `position_updated` | Position modified by order fill |
| `position_lost` | Market resolved against position |
| `payout_claimed` | Winnings withdrawn (also used when a position wins) |

### History Event Fields

Each history event includes these fields. All USD values are in micro USD (1,000,000 = \$1.00).

| Field | Type | Description |
| :- | :- | :- |
| `id` | integer | Unique history event identifier |
| `eventType` | string | Event type (see table above) |
| `signature` | string | Solana transaction signature |
| `slot` | string | Solana slot number |
| `timestamp` | integer | Unix timestamp (seconds) when the event occurred |
| `orderPubkey` | string | Associated order account public key |
| `positionPubkey` | string | Associated position account public key |
| `marketId` | string | External market identifier |
| `ownerPubkey` | string | Position or order owner public key |
| `keeperPubkey` | string | Keeper that processed the transaction |
| `externalOrderId` | string | Client-provided order identifier |
| `orderId` | string | External order identifier from the venue |
| `isBuy` | boolean | `true` when the event relates to a buy order |
| `isYes` | boolean | `true` when the event relates to the YES side |
| `contracts` | string | Number of contracts in the order |
| `filledContracts` | string | Number of contracts filled |
| `contractsSettled` | string | Number of contracts settled in this event |
| `avgFillPriceUsd` | string | Average fill price in micro USD |
| `maxFillPriceUsd` | string | Maximum fill price in micro USD |
| `maxBuyPriceUsd` | string, nullable | Buyer-specified max fill price in micro USD |
| `minSellPriceUsd` | string, nullable | Seller-specified min fill price in micro USD |
| `depositAmountUsd` | string | Amount deposited for this order in micro USD |
| `totalCostUsd` | string | Total cost of the order in micro USD |
| `feeUsd` | string, nullable | Fee charged for this event in micro USD |
| `grossProceedsUsd` | string | Gross proceeds before fees in micro USD |
| `netProceedsUsd` | string | Net proceeds after fees in micro USD |
| `transferAmountToken` | string, nullable | Token amount transferred (native token units) |
| `realizedPnl` | string, nullable | Realized P\&L in micro USD |
| `realizedPnlBeforeFees` | string, nullable | Realized P\&L before fees in micro USD |
| `payoutAmountUsd` | string | Payout amount in micro USD (populated for `payout_claimed` events) |
| `eventId` | string | External event identifier |
| `marketMetadata` | object | Metadata for the associated market |
| `eventMetadata` | object | Metadata for the parent event |

### Get Transaction History

```js theme={null}
const response = await fetch(
  'https://api.jup.ag/prediction/v1/history?' + new URLSearchParams({
    ownerPubkey: wallet.publicKey.toString()
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const history = await response.json();
```

***

## What's Next

Explore social features to view profiles, follow traders, and track leaderboards.

> [card] [Social Features](https://developers.jup.ag/docs/prediction/social-features.md)
>
> Profiles, leaderboards, and trade feeds


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.