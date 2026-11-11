> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Prediction Market API

> Jupiter's Prediction Market API: discover events, read prices, trade YES/NO contracts and claim payouts on Solana prediction markets at api.jup.ag/prediction/v1.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

The Prediction Market API is Jupiter's REST API for prediction markets on Solana, served at `https://api.jup.ag/prediction/v1`. A prediction market lets people trade on the outcome of a real-world event: each market has a YES and a NO side, a contract on the winning side pays \$1, and the price of a contract reflects the probability the market gives that outcome. The API covers the whole life of a trade:

* Discover events and their binary markets across sports, crypto, politics, e-sports, culture, economics and tech, with live prices and orderbook depth.
* Build buy and sell orders as unsigned Solana transactions, sign them with the user's wallet, and submit them through `POST /execute` or your own RPC.
* Track positions, orders, P\&L and history, and claim payouts of \$1 per winning contract, with no claim fee.

Orders are funded in JupUSD or USDC. Trading applications, wallets, bots and analytics dashboards are the typical integrators.

> [card] [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md)
>
> Discover events and get market details

> [card] [Open Positions](https://developers.jup.ag/docs/prediction/open-positions.md)
>
> Create buy orders for YES/NO contracts

> [card] [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md)
>
> View, sell, and manage your holdings

> [card] [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md)
>
> Claim winnings after settlement

## Quick start

List one trending event with its markets. No API key is needed:

```bash Request theme={null}
curl "https://api.jup.ag/prediction/v1/events?filter=trending&includeMarkets=true&start=0&end=1"
```

```json Response (trimmed) theme={null}
{
  "data": [
    {
      "eventId": "POLY-45915",
      "category": "politics",
      "metadata": { "title": "Brazil Presidential Election" },
      "markets": [
        {
          "marketId": "POLY-601826",
          "title": "Flávio Bolsonaro",
          "status": "open",
          "pricing": {
            "buyYesPriceUsd": 850000,
            "sellYesPriceUsd": 849000,
            "buyNoPriceUsd": 151000,
            "sellNoPriceUsd": 150000
          }
        }
      ]
    }
  ]
}
```

Prices are in micro USD: `850000` is \$0.85, an implied probability of 85% for YES. Take a `marketId` from the response into `POST /orders` to build a buy order; the full sequence is on [Trading Lifecycle](https://developers.jup.ag/docs/prediction/trading-lifecycle.md).

> [!TIP]
>
> Keyless requests are limited to 0.5 requests per second. For production, pass an API key in the `x-api-key` header. See [Rate Limits](https://developers.jup.ag/docs/portal/rate-limits.md).

## What prediction markets are

A prediction market is a market on the outcome of a real-world occurrence, such as an election result, a match score or a token price at a date. On Jupiter every market is binary:

* **Two outcomes.** Each market has a YES side and a NO side. An event groups one or more markets, for example one market per candidate or per team.
* **Contracts pay \$1.** A contract on the winning side is worth \$1 at settlement; a contract on the losing side is worth nothing. Buying YES at \$0.70 and winning returns \$0.30 of profit per contract.
* **Price is probability.** A YES price of \$0.70 means the market gives the outcome a 70% chance, and YES plus NO prices add up to about \$1.
* **Settlement is final.** When the event resolves, the result is recorded onchain, winning positions become claimable and payouts come from the vault, with no claim fee. [Jupiter Forecast](https://developers.jup.ag/docs/prediction/forecast.md) rounds are the exception: they settle atomically, so there is no claim step.

The sections [Architectural Components](#architectural-components) and [Key Concepts](#key-concepts) below describe events, markets, orders, positions, settlement, fees and fractional contracts in detail.

## What the API does

| Area | Endpoints | Documentation |
| - | - | - |
| Discover events and markets | `GET /events`, `GET /events/search`, `GET /events/{eventId}`, `GET /events/{eventId}/markets`, `GET /events/{eventId}/markets/{marketId}`, `GET /events/suggested/{pubkey}`, `GET /markets/{marketId}` | [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md) |
| Read prices, depth and live scores | `GET /orderbook/{marketId}`, `GET /trading-status`, `GET /events/scores`, `GET /events/{eventId}/score` | [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md#get-orderbook-data) |
| Open positions | `POST /orders` (unsigned buy transaction), `POST /execute` (submit the signed transaction) | [Open Positions](https://developers.jup.ag/docs/prediction/open-positions.md) |
| Manage positions | `DELETE /positions/{positionPubkey}` and `DELETE /positions` to sell, `GET /orders`, `GET /orders/{orderPubkey}`, `GET /orders/status/{orderPubkey}` | [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md) |
| Track activity | `GET /positions`, `GET /positions/{positionPubkey}`, `GET /history` | [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md) |
| Claim payouts | `POST /positions/{positionPubkey}/claim` | [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md) |
| Profiles and social | `GET /profiles/{ownerPubkey}`, `GET /profiles/{ownerPubkey}/pnl-history`, `GET /trades`, `GET /leaderboards` | [Social Features](https://developers.jup.ag/docs/prediction/social-features.md) |
| Vault and forecast series | `GET /vault-info`, `GET /forecast?marketId=` (numeric forecast time series for a market) | [API Reference](https://developers.jup.ag/docs/api-reference/prediction/get-vault-info.md) |

Every endpoint is specified in the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction/get-events.md).

## How a trade works

**1. Discover an event**

`GET /events` with `category`, `filter` or `tags`, or `GET /events/search?query=` for a keyword. Each event carries a `markets[]` array; pick the `marketId` of the outcome the user is trading.

**2. Read the price**

`GET /markets/{marketId}` returns fresh `buyYesPriceUsd`, `buyNoPriceUsd`, `sellYesPriceUsd` and `sellNoPriceUsd` in micro USD. Re-fetch right before ordering; `GET /orderbook/{marketId}` shows how much size fills near that price.

**3. Build the order**

`POST /orders` with `ownerPubkey`, `marketId`, `isYes`, `isBuy: true`, `depositAmount` and `depositMint` (JupUSD or USDC). The minimum order is \$5. The response is an unsigned Solana transaction.

**4. Sign and execute**

Sign the transaction with the user's wallet and submit it, either to `POST /execute` with the build's `execution.context` or through your own RPC. Jupiter's keeper network fills the order on the underlying market.

**5. Track the position**

`GET /orders/status/{orderPubkey}` reports `created`, `partiallyfilled`, `filled` or `failed`; `GET /positions?ownerPubkey=` returns contracts held, cost basis, mark-to-market value and P\&L.

**6. Sell or claim**

Sell before resolution with `DELETE /positions/{positionPubkey}`, or hold to settlement and call `POST /positions/{positionPubkey}/claim` to receive \$1 per winning contract.

Full code for each step, with the data that flows from one call to the next, is on [Trading Lifecycle](https://developers.jup.ag/docs/prediction/trading-lifecycle.md). For a complete application with video walkthrough and source, see [How to Build a Prediction Market App on Solana](https://developers.jup.ag/docs/guides/how-to-build-a-prediction-market-app-on-solana.md).

## Typical integrations

> [card] [Trading apps and wallets](https://developers.jup.ag/docs/guides/how-to-build-a-prediction-market-app-on-solana.md)
>
> Event discovery, order building and position tracking behind your own interface, with the user signing in their wallet.

> [card] [Bots and market participants](https://developers.jup.ag/docs/prediction/events-and-markets.md#get-orderbook-data)
>
> Orderbook depth, trading status and order status polling for automated strategies.

> [card] [Dashboards and social products](https://developers.jup.ag/docs/prediction/social-features.md)
>
> Profiles, P\&L history, leaderboards and the global trade feed.

> [card] [Short-term BTC markets](https://developers.jup.ag/docs/prediction/forecast.md)
>
> Jupiter Forecast's 15-minute BTC up/down rounds, tradable through the Prediction API or directly as tokens through the Swap API.

## Geographical Restrictions

* We have restricted United States and South Korea IPs from accessing the Prediction Market API.
* If you have any questions, reach out to us in Discord/Telegram.

## Architectural Components

We have generalised the Prediction Market architecture to be agnostic of the underlying data provider. This allows us to support multiple data providers and add new ones in the future.

### Events

Events are the top-level container for prediction markets. Each event represents a real-world occurrence that will have a determinable outcome.

**Event Properties:**

* **Event ID**: Unique identifier for the event
* **Series**: Groups related events (e.g., "NFL Week 1")
* **Category**: Primary classification (crypto, sports, politics, esports, culture, economics, tech)
* **Subcategory**: More specific classification within the category
* **Markets**: One or more binary markets within the event
* **Metadata**: Title, subtitle, images, rules documents
* **Metrics**: Total TVL, trading volume, close conditions

**Event Lifecycle:**

* Events can contain multiple markets (e.g., "Who will win the championship?" might have markets for each team)
* Events can support single winners or multiple winners depending on configuration
* Events track overall volume and liquidity across all their markets

### Markets

Markets are binary prediction markets within an event, representing specific YES/NO outcomes.

**Market Properties:**

* **Market ID**: Unique identifier derived as a PDA (Program Derived Address)
* **Status**: `open`, `closed`, `cancelled`
* **Result**: `null` (not yet resolved), `yes`, or `no`
* **Pricing Data**: Buy/sell prices for YES and NO sides, volume, open interest, liquidity
* **Metadata**: Title, description, trading rules, resolution criteria

**Market States:**

1. **Open**: Market is actively trading, users can create orders
2. **Closed**: Trading has stopped, awaiting settlement

**Pricing Model:**

* Prices are quoted in USD (stored as micro USD on-chain)
* Each contract represents a claim to \$1 if the outcome is correct
* Separate buy and sell prices for YES and NO sides
* Prices reflect probability (e.g., 70¢ YES price = 70% probability)

### Orders

Orders represent user intentions to buy or sell contracts in a market.

**Order Properties:**

* **Order Pubkey**: Solana account address for the order
* **Owner**: User's public key
* **Market**: Target market identifier
* **Side**: YES or NO position
* **Direction**: Buy
* **Contracts**: Number of contracts to trade
* **Status**: `pending`, `filled`, or `failed`
* **Fill Details**: Filled contracts, average fill price, fees paid

**Order Flow:**

1. User creates an order request with desired parameters
2. API returns an unsigned Solana transaction
3. User signs and submits the transaction to Solana
4. Keeper network matches and fills the order
5. Order status updates reflect fill progress
6. Settled orders update associated positions

### Positions

Positions represent a user's holdings in a specific market side (YES or NO).

**Position Properties:**

* **Position Pubkey**: PDA derived from owner, market, and side
* **Owner**: User's public key
* **Market**: Associated market identifier
* **Side**: YES or NO position
* **Contracts**: Total contracts held
* **Cost Basis**: `totalCostUsd`, `avgPriceUsd`
* **Mark-to-Market**: `valueUsd`, `markPriceUsd` (current market value)
* **P\&L Metrics**: Unrealised P\&L, P\&L after fees, percentage returns
* **Order Activity**: Number of open orders
* **Settlement**: `claimable`, `claimed`, `claimedUsd`, `payoutUsd`

**Position States:**

* **Active**: Market is open, position has mark-to-market value and P\&L
* **Closed Market**: Market closed but not settled, P\&L frozen
* **Claimable**: Market settled favourably, payout available
* **Lost**: Market settled against the position (no payout)
* **Claimed**: Payout has been withdrawn

**P\&L Calculation:**

* **Unrealised P\&L**: (Current market value) - (Cost basis)
* **Realised P\&L**: Calculated on partial or full position closes
* **Fees**: Protocol and venue fees reduce net returns

### Vault

The Vault manages the settlement and escrow mechanism for the prediction market system.

**Vault Functions:**

* **Escrow Management**: Holds deposited JupUSD or USDC during active trading
* **Settlement Distribution**: Pays out winning positions
* **Fee Collection**: Accumulates protocol and venue fees
* **Balance Tracking**: Maintains total locked liquidity

The vault is implemented as a Solana PDA that securely holds all user deposits and ensures payouts are available when positions are claimable.

### History & Events

The History system tracks all on-chain activity for audit and user interface purposes.

**Event Types:**

* `order_created`: New order placed
* `order_filled`: Order matched and executed
* `order_failed`: Order could not be filled
* `order_closed`: Order account closed after completion
* `position_updated`: Position modified by order fill
* `position_lost`: Position resolved against user
* `payout_claimed`: User withdrew winnings

**History Properties:**

* Signature and slot for on-chain verification
* Full order and position details at time of event
* Fee breakdown and P\&L calculations
* Market and event metadata snapshots

### Orderbook & Price Discovery

The orderbook system facilitates price discovery and order matching.

**Orderbook Structure:**

* **Bid/Ask Spreads**: Separate order books for YES and NO sides
* **Liquidity Depth**: Available contracts at various price levels
* **Market Data**: Best bid/ask, spreads, recent trades
* **Volume Tracking**: 24h volume, all-time volume

**Price Discovery:**

* Prices are determined by supply and demand
* Market makers can provide liquidity at various price points
* Keepers match orders efficiently
* Slippage protection via price limits

### Settlement Process

The settlement process resolves markets and enables winners to claim payouts.

**Settlement Flow:**

1. **Market Closes**: Trading stops at predetermined `closeTime`
2. **Result Determined**: Authoritative source provides outcome
3. **Market Settles**: Result recorded on-chain (`yes` or `no`)
4. **Positions Become Claimable**: Winning positions can claim payouts
5. **Users Claim**: Winners receive \$1 per contract (no claim fees)
6. **Losing Positions**: Contracts become worthless

**Settlement Guarantees:**

* All outcomes are determined by transparent rules
* Settlement is final and immutable once recorded
* Payouts are guaranteed by vault reserves
* Disputed outcomes follow documented resolution procedures

## Key Concepts

### Binary Market Structure

Every market has exactly two outcomes: YES or NO. This simplicity ensures:

* Clear resolution criteria
* Straightforward pricing (probability-based)
* Simple position management
* Transparent payouts

### Contract Value

Each contract represents a claim to **\$1 if correct**:

* If you buy YES at 70¢ and YES wins, you profit 30¢ per contract
* If you buy NO at 40¢ and NO wins, you profit 60¢ per contract
* Losing contracts expire worthless

### Fractional Contracts

Contracts are fractional. Each response carries three forms of every contract quantity:

* `*Micro` (e.g. `contractsMicro`): exact amount in micro-contracts, where `1000000` = 1 contract. Use this for any math.
* `*Decimal` (e.g. `contractsDecimal`): exact amount as a decimal string (e.g. `"5.777773"`). Use this for display.
* legacy whole-contract field (e.g. `contracts`): the amount **floored to a whole number**. Kept for backwards compatibility only; fills can be fractional, so this value is truncated and must not be used for accounting.

This applies across the API: `contracts`, `filledContracts`, `contractsSettled`, `newContracts`, and `totalActiveContracts` each have `*Micro` and `*Decimal` siblings.

> [!WARNING]
>
> `*Micro` values are returned as strings (`u64`/`u128`) and can exceed JavaScript's safe-integer range. Parse them with `BigInt` or a decimal library, never `parseFloat` or `Number`.

When creating an order you may send `contractsMicro` (or `contractsDecimal`, or the legacy `contracts`), but only one quantity field per request. Effective fill sizes are floored to 0.01-contract increments; you can submit finer sizes, but executed amounts round down to the nearest 0.01.

### Price as Probability

Market prices reflect implied probability:

* 70¢ YES price ≈ 70% probability of YES outcome
* 30¢ NO price ≈ 30% probability of NO outcome
* YES price + NO price ≈ \$1.00

### Position Management

Users can actively manage positions:

* **Buy more**: Increase exposure to an outcome
* **Close position**: Exit the market completely
* **Hold to settlement**: Wait for final payout

### Fees

These are the breakdown of fees:

* Fees are only charged on executed trades - buying or selling of contracts.
* There are no fees when claiming payouts.
* Fees are charged in whichever mint is used to purchase the contracts.
* Fees are always rounded up to the nearest cent.
* Fees are taken only on executed trades (Orders that do not fill immediately, you pay no trading fee until it later executes).

Your trading fees depend on three factors:

* The price of the contract
* How many contracts you trade (Bigger trades = higher fees)
* How uncertain the outcome is (More uncertainty = higher fees)
  * Trades priced closer to \$0.50 are more uncertain, so they have slightly higher fees.
  * Trades priced closer to \$0.00 or \$1.00 are more certain, so fees are lower.

Here's what fees look like for common prices:

| Price per Contract | Fee (1 Contract) | Fee (100 Contracts) |
| :-: | :-: | :-: |
| \$0.01 | \$0.01 | \$0.07 |
| \$0.05 | \$0.01 | \$0.34 |
| \$0.10 | \$0.01 | \$0.63 |
| \$0.15 | \$0.01 | \$0.90 |
| \$0.20 | \$0.02 | \$1.12 |
| \$0.25 | \$0.02 | \$1.32 |
| \$0.30 | \$0.02 | \$1.47 |
| \$0.35 | \$0.02 | \$1.60 |
| \$0.40 | \$0.02 | \$1.68 |

Example:

* If you buy 100 contracts at \$0.25:
* Total cost: \$25.00
* Trading fee: \$1.32

Why Fees Work This Way:

* Keeps fees low for high-confidence outcomes.
* Scales fairly with trade size.
* Encourages liquidity by not charging resting orders.

### Account Structure

There are a few accounts involved in the prediction market system:

1. **Position Account**: Represents a user's position in a market.
2. **Order Account**: Represents a user's order to open, close, or modify a position.
3. **Vault Account**: Represents the vault that holds the user's funds.

Upon position creation, there will be 3 transactions involved:

1. Create order: When opening a position, you sign a transaction to create an order for the position.
2. Fill order: Once the order is created, our keeper system will pick up the order and fill them directly on the underlying prediction markets.
3. Close order: Once the order is filled or not, the order will be closed.

## FAQ

> [accordion] What endpoints are included in the Prediction Market API?
>
> Discovery: `/events`, `/events/search`, `/events/{eventId}`, `/events/{eventId}/markets`, `/events/{eventId}/markets/{marketId}`, `/events/suggested/{pubkey}`, `/markets/{marketId}`, `/orderbook/{marketId}`, `/events/scores`, `/events/{eventId}/score` and `/trading-status`. Trading: `POST /orders` to build a buy order, `DELETE /positions/{positionPubkey}` or `DELETE /positions` to sell, and `POST /execute` to submit a signed transaction. Tracking: `/orders`, `/orders/{orderPubkey}`, `/orders/status/{orderPubkey}`, `/positions`, `/positions/{positionPubkey}` and `/history`. Settlement: `POST /positions/{positionPubkey}/claim`. Social: `/profiles/{ownerPubkey}`, `/profiles/{ownerPubkey}/pnl-history`, `/trades` and `/leaderboards`. Plus `/vault-info` and `/forecast?marketId=`, a numeric forecast time series for a market. The table in [What the API does](#what-the-api-does) links each group to its guide; the [API Reference](https://developers.jup.ag/docs/api-reference/prediction/get-events.md) has every parameter.

> [accordion] How do I query active prediction market orders via API?
>
> Call `GET /orders?ownerPubkey=<wallet>` to list a wallet's orders, each with its `status` (`pending`, `filled` or `failed`), filled contracts, average fill price and fees. For a single order, `GET /orders/status/{orderPubkey}` returns the status alone (`created`, `partiallyfilled`, `filled` or `failed`), which is what you poll after submitting. `GET /orders/{orderPubkey}` returns the full order while its account exists and `400` once a filled order has been closed, so use `/history` for past orders. Open positions rather than orders come from `GET /positions?ownerPubkey=<wallet>`. See [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md#querying-orders).

> [accordion] How does a trade reach the blockchain?
>
> The API never holds keys. `POST /orders` returns an unsigned Solana transaction; your application signs it with the user's wallet and submits it, through `POST /execute` with the build's `execution.context` or through your own RPC. Jupiter's keeper network then fills the order on the underlying market and the order status moves from `created` to `filled` or `failed`. Selling works the same way from `DELETE /positions/{positionPubkey}`; a claim transaction from `/claim` is signed and sent to Solana directly.

> [accordion] Which currency are trades settled in?
>
> Orders are funded in JupUSD or USDC (`depositMint` on `POST /orders`), with a minimum of \$5 per order. Fees are charged in the mint used to buy. Payouts are \$1 per winning contract. Jupiter Forecast markets accept USDC only.

> [accordion] What fees does the Prediction Market API charge?
>
> Fees apply to executed buys and sells only: nothing on resting orders and nothing on claims. The fee grows with trade size and with uncertainty, so contracts priced near \$0.50 cost slightly more than contracts near \$0.01 or \$0.99, and every fee is rounded up to the cent. The table in [Fees](#fees) gives the amounts for common prices; `POST /orders` responses include the fee for that order.

> [accordion] How do I retrieve active prediction markets and events?
>
> `GET /events?filter=live` returns events in progress and `filter=trending`, `new` or `upcoming` the other states; add `includeMarkets=true` to get each event's markets inline, or `category` and `tags` to narrow the list. A market is tradable while its `status` is `open`; `GET /trading-status` tells you whether trading is currently enabled. See [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md).

> [accordion] Can users close positions before an event resolves?
>
> Yes, while the market is `open`. `DELETE /positions/{positionPubkey}` builds a sell transaction for a position and `DELETE /positions` for all of a wallet's positions; sign and submit it like a buy. Sales are executed trades, so the trading fee applies. See [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md).

> [accordion] Can I retrieve trading history and P\&L data?
>
> Yes. `GET /positions?ownerPubkey=` returns each position with cost basis, mark-to-market value and P\&L fields, `GET /history` returns the full event trail (orders created, filled, failed and closed, positions updated or lost, payouts claimed) with signatures and slots, and `GET /profiles/{ownerPubkey}/pnl-history` returns a wallet's P\&L over time. See [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md) and [Social Features](https://developers.jup.ag/docs/prediction/social-features.md).

> [accordion] Is the API available in the United States?
>
> No. United States and South Korea IP addresses are blocked from the Prediction Market API. See [Geographical Restrictions](#geographical-restrictions).

> [accordion] Do I need an API key?
>
> Not to explore: keyless requests work at 0.5 requests per second. For production, pass an API key in the `x-api-key` header; one key covers every Jupiter API. See [Rate Limits](https://developers.jup.ag/docs/portal/rate-limits.md).

> [accordion] How are contract quantities represented?
>
> Contracts are fractional. Every quantity comes in three forms: `*Micro` as a string in micro-contracts (`1000000` = 1 contract, use it for math and parse it with `BigInt`), `*Decimal` as a decimal string for display, and a legacy whole-number field that is floored and must not be used for accounting. See [Fractional Contracts](#fractional-contracts).

> [accordion] How do winners get paid?
>
> When an event resolves, the result is recorded onchain and winning positions become `claimable`. `POST /positions/{positionPubkey}/claim` builds the claim transaction, which pays \$1 per contract from the vault with no claim fee. Some markets settle automatically without a claim, and a claimable payout that is not claimed within 24 hours is claimed on the winner's behalf by the keeper network, with no fee; [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md) describes those cases and how to detect a lost position.

> [accordion] What is Jupiter Forecast?
>
> Jupiter's native prediction market on Solana. The current markets are 15-minute Bitcoin up/down rounds settled from the Chainlink BTC/USD price. Each side is a Token-2022 mint worth \$1 if it wins, so the rounds can be traded through the Prediction API, which tracks positions and P\&L for you, or directly through the Swap API as token swaps. See [Jupiter Forecast](https://developers.jup.ag/docs/prediction/forecast.md).


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.