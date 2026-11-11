> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Order (Trigger)

> Create price-triggered limit orders: single, OCO (take-profit/stop-loss), and OTOCO (conditional chains).

Price orders execute a swap when a token's USD price crosses a threshold: limit orders, take-profit, and stop-loss. This page covers the create call and its order types. The vault and deposit setup is shared with DCA: the example below includes it inline, and [Vault & Deposit](https://developers.jup.ag/docs/trigger/deposit.md) is the full reference.

> [!TIP]
>
> For a working reference of the full order creation flow, see how [jup.ag](https://jup.ag) handles it. The frontend implements the same API with validation, error handling, and UX patterns you can use as a guide.

## Order types

Set `orderType` on the create call to choose the variant.

| Type | Behaviour |
| :- | :- |
| `single` | Triggers a swap when `triggerMint`'s USD price crosses above or below a target. Standard limit order, stop-loss, or [trailing stop loss](#trailing-stop-loss). |
| `oco` | A take-profit and stop-loss pair sharing one deposit. When one side fills, the other cancels automatically. |
| `otoco` | A parent trigger that executes first, then activates a TP/SL (OCO) pair on the output. |

> [!NOTE]
>
> On the create call, `orderType` is the variant (`single`, `oco`, `otoco`). That is different from the deposit's `orderType` (`price` or `dca`). Craft the deposit with `orderType: "price"` and an `orderSubType` that matches the create `orderType`.

## Quick start

Creating a price order is four calls: authenticate, get your vault, craft and sign the deposit, then submit the order.

> [accordion] Prerequisites
>
> Install the signing dependencies:
>
> ```bash theme={null}
> npm install @solana/web3.js bs58 tweetnacl
> ```
>
> You also need a [Jupiter API key](https://developers.jup.ag/portal) and a funded wallet. The example below loads the wallet from `BS58_PRIVATE_KEY` and the key from `JUPITER_API_KEY` in your `.env`. For the browser-wallet (`signMessage` / `signTransaction`) version of authentication and signing, see [Authentication](https://developers.jup.ag/docs/trigger/authentication.md) and [Vault & Deposit](https://developers.jup.ag/docs/trigger/deposit.md).
>
> > [!WARNING]
> >
> > Never commit private keys. Use environment variables for testing and a proper key-management solution in production.

```typescript expandable theme={null}
import { Keypair, VersionedTransaction } from "@solana/web3.js";
import bs58 from "bs58";
import nacl from "tweetnacl";

const BASE = "https://api.jup.ag/trigger/v2";
const API_KEY = process.env.JUPITER_API_KEY!;
const SOL = "So11111111111111111111111111111111111111112";
const USDC = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v";

const wallet = Keypair.fromSecretKey(bs58.decode(process.env.BS58_PRIVATE_KEY!));
const owner = wallet.publicKey.toBase58();

// 1. Authenticate: sign the challenge to get a 24h JWT (see /trigger/authentication)
async function authenticate(): Promise<string> {
  const { challenge } = await fetch(`${BASE}/auth/challenge`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "x-api-key": API_KEY },
    body: JSON.stringify({ walletPubkey: owner, type: "message" }),
  }).then((r) => r.json());

  const signature = nacl.sign.detached(new TextEncoder().encode(challenge), wallet.secretKey);
  const { token } = await fetch(`${BASE}/auth/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "x-api-key": API_KEY },
    body: JSON.stringify({ type: "message", walletPubkey: owner, signature: bs58.encode(signature) }),
  }).then((r) => r.json());
  return token;
}

async function createLimitOrder() {
  const token = await authenticate();
  const headers = { "Content-Type": "application/json", "x-api-key": API_KEY, Authorization: `Bearer ${token}` };

  // 2. Get your vault, registering one on first use
  let vault = await fetch(`${BASE}/vault`, { headers });
  if (!vault.ok) vault = await fetch(`${BASE}/vault/register`, { headers });
  if (!vault.ok) throw new Error(`vault failed: ${vault.status}`);

  // 3. Craft the deposit (orderType: "price", orderSubType matches the create orderType below)
  const deposit = await fetch(`${BASE}/deposit/craft`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      inputMint: SOL,
      outputMint: USDC,
      userAddress: owner,
      amount: "1000000000", // 1 SOL
      orderType: "price",
      orderSubType: "single",
    }),
  }).then((r) => r.json());
  if (!deposit.requestId) throw new Error(`craft failed: ${JSON.stringify(deposit)}`);

  // 4. Sign the deposit and submit the order
  const tx = VersionedTransaction.deserialize(Buffer.from(deposit.transaction, "base64"));
  tx.sign([wallet]);
  const depositSignedTx = Buffer.from(tx.serialize()).toString("base64");

  const order = await fetch(`${BASE}/orders/price`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      orderType: "single",
      depositRequestId: deposit.requestId,
      depositSignedTx,
      userPubkey: owner,
      inputMint: SOL,
      outputMint: USDC,
      inputAmount: "1000000000",
      triggerMint: SOL,
      triggerCondition: "above", // fill when SOL climbs above the target
      triggerPriceUsd: 200,
      slippageBps: 100,          // 1%
      expiresAt: Date.now() + 7 * 24 * 60 * 60 * 1000, // 7 days
    }),
  }).then((r) => r.json());
  if (!order.id) throw new Error(`create failed: ${JSON.stringify(order)}`);

  console.log(`Order ${order.id}, deposit tx https://solscan.io/tx/${order.txSignature}`);
  return order.id as string;
}

createLimitOrder().catch(console.error);
```

The deposit lands on-chain during the create call. The response is `{ id, txSignature, depositConfirmed }`: `id` is the order identifier you use to [track](https://developers.jup.ag/docs/trigger/order-history.md), [update, or cancel](https://developers.jup.ag/docs/trigger/manage-orders.md) it, `txSignature` is the on-chain deposit signature, and `depositConfirmed` is `true` once the deposit has landed and the order is active:

```json theme={null}
{
  "id": "order-uuid",
  "txSignature": "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d...",
  "depositConfirmed": true
}
```

> [!NOTE]
>
> The vault address is resolved from your JWT, so you never pass it. The deposit `requestId` is single-use and is consumed by the create call as `depositRequestId`.

## Other order types

The OCO and OTOCO bodies replace step 4's create call. Each needs its deposit crafted with the matching `orderSubType`.

### OCO (One-Cancels-Other)

An OCO order creates a take-profit and stop-loss pair sharing one deposit. When one side fills, the other cancels automatically.

```typescript theme={null}
// deposit crafted with orderSubType: "oco"
const order = await fetch(`${BASE}/orders/price`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    orderType: "oco",
    depositRequestId: deposit.requestId,
    depositSignedTx,
    userPubkey: owner,
    inputMint: SOL,
    outputMint: USDC,
    inputAmount: "1000000000",
    triggerMint: SOL,
    tpPriceUsd: 250, // take profit
    slPriceUsd: 150, // stop loss
    tpSlippageBps: 100,
    slSlippageBps: 100,
    expiresAt: Date.now() + 7 * 24 * 60 * 60 * 1000,
  }),
}).then((r) => r.json());
```

> [!NOTE]
>
> The take-profit price must be greater than the stop-loss price.

### OTOCO (One-Triggers-One-Cancels-Other)

An OTOCO order has a parent trigger that executes first. Once filled, it automatically activates a TP/SL pair (OCO) on the output tokens.

```typescript theme={null}
// deposit crafted with orderSubType: "otoco"
const order = await fetch(`${BASE}/orders/price`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    orderType: "otoco",
    depositRequestId: deposit.requestId,
    depositSignedTx,
    userPubkey: owner,
    inputMint: USDC,
    outputMint: SOL,
    inputAmount: "200000000", // 200 USDC
    triggerMint: SOL,
    triggerCondition: "below",
    triggerPriceUsd: 180, // entry: buy SOL when it drops to \$180
    tpPriceUsd: 220,      // take profit at \$220
    slPriceUsd: 160,      // stop loss at \$160
    slippageBps: 100,     // parent order slippage
    tpSlippageBps: 100,
    slSlippageBps: 100,
    expiresAt: Date.now() + 30 * 24 * 60 * 60 * 1000, // 30 days
  }),
}).then((r) => r.json());
```

## Trailing stop loss

A trailing stop loss is a `single` order that tracks the market instead of using a fixed price. Set `trailingBps` instead of `triggerPriceUsd`: the trigger follows the price in your favour and holds when the price reverses. Provide exactly one of `triggerPriceUsd` (static) or `trailingBps` (trailing).

`trailingBps` is the trail distance in basis points, from 50 to 9000 (0.5% to 90%). `triggerCondition` sets the direction:

| `triggerCondition` | Use | `triggerMint` | Trigger price |
| :- | :- | :- | :- |
| `below` | Sell-below (stop loss): protect a position, sell if the price falls | `inputMint` | `highWatermark × (1 − trailingBps/10000)` |
| `above` | Buy-above: enter as the price rises off a low | `outputMint` | `lowWatermark × (1 + trailingBps/10000)` |

At create, the API seeds the trigger from the current market price. The keeper then tracks a running watermark (the high for `below`, the low for `above`) and rewrites the trigger as the watermark moves in your favour. The trigger never moves against you. Watermark updates are batched, not per price tick.

Craft the deposit with `orderSubType: "single"`, then submit a single order with `trailingBps` and no `triggerPriceUsd`:

```typescript theme={null}
// deposit crafted with orderSubType: "single"; hold SOL, sell if it falls 10% from its peak
const order = await fetch(`${BASE}/orders/price`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    orderType: "single",
    depositRequestId: deposit.requestId,
    depositSignedTx,
    userPubkey: owner,
    inputMint: SOL,           // selling the asset you hold
    outputMint: USDC,
    inputAmount: "200000000", // 0.2 SOL
    triggerMint: SOL,         // "below" requires triggerMint === inputMint
    triggerCondition: "below",
    trailingBps: 1000,        // trail 10% below the peak; no triggerPriceUsd
    slippageBps: 100,
    expiresAt: Date.now() + 30 * 24 * 60 * 60 * 1000,
  }),
}).then((r) => r.json());
```

`slippageBps` is the execution buffer beneath the trigger, the same as a static order. Track the live trigger and watermark with [Order History](https://developers.jup.ag/docs/trigger/order-history.md) (`triggerPriceUsd`, `trailingBps`, `highWatermark`/`lowWatermark`), and change the trail distance later by [updating](https://developers.jup.ag/docs/trigger/manage-orders.md) `trailingBps`.

> [!NOTE]
>
> Trailing rules: provide exactly one of `triggerPriceUsd` or `trailingBps` (else `400 "exactly one of triggerPriceUsd or trailingBps must be set"`); `trailingBps` must be 50–9000; and `triggerMint` must be the `inputMint` for `below` or the `outputMint` for `above`.

## Order parameters reference

### Common fields (all order types)

| Parameter | Type | Required | Description |
| :- | :- | :- | :- |
| `orderType` | `string` | Yes | `single`, `oco`, or `otoco` |
| `depositRequestId` | `string` | Yes | The `requestId` from the deposit craft response |
| `depositSignedTx` | `string` | Yes | Base64-encoded signed deposit transaction |
| `userPubkey` | `string` | Yes | Your wallet public key. Must match the JWT |
| `inputMint` | `string` | Yes | Mint address of the token to sell |
| `inputAmount` | `string` | Yes | Amount in smallest unit (lamports, etc.) |
| `outputMint` | `string` | Yes | Mint address of the token to buy |
| `triggerMint` | `string` | Yes | Mint address to monitor for the price trigger |
| `expiresAt` | `number` | Yes | Expiration timestamp in milliseconds |

### Single order fields

| Parameter | Type | Required | Description |
| :- | :- | :- | :- |
| `triggerCondition` | `string` | Yes | `above` or `below` |
| `triggerPriceUsd` | `number` | Cond. | USD price that triggers execution. Required unless `trailingBps` is set. |
| `trailingBps` | `number` | Cond. | Trail distance in basis points (50–9000) for a [trailing stop loss](#trailing-stop-loss). Set instead of `triggerPriceUsd`. |
| `slippageBps` | `number` | No | Slippage tolerance in basis points (0-10000) |

### OCO order fields

| Parameter | Type | Required | Description |
| :- | :- | :- | :- |
| `tpPriceUsd` | `number` | Yes | Take-profit trigger price (USD) |
| `slPriceUsd` | `number` | Yes | Stop-loss trigger price (USD) |
| `tpSlippageBps` | `number` | No | Take-profit slippage (basis points) |
| `slSlippageBps` | `number` | No | Stop-loss slippage (basis points) |

### OTOCO order fields

| Parameter | Type | Required | Description |
| :- | :- | :- | :- |
| `triggerCondition` | `string` | Yes | Parent trigger: `above` or `below` |
| `triggerPriceUsd` | `number` | Yes | Parent trigger price (USD) |
| `tpPriceUsd` | `number` | Yes | Take-profit price after parent fills |
| `slPriceUsd` | `number` | Yes | Stop-loss price after parent fills |
| `slippageBps` | `number` | No | Parent order slippage |
| `tpSlippageBps` | `number` | No | Take-profit slippage |
| `slSlippageBps` | `number` | No | Stop-loss slippage |

## Default slippage

If slippage is not specified, defaults vary by order type and trigger condition:

| Order side | Default slippage |
| :- | :- |
| **Take-profit / buy below** (buying into strength) | Auto slippage via RTSE (Real-Time Slippage Estimator) |
| **Stop-loss / buy above** (selling into weakness) | 20% (2000 bps) |

Stop-loss orders use a higher default because execution certainty is more important than price precision when cutting losses.

## Validation rules

* Minimum order size: 10 USD
* Input and output mints must be different
* Slippage: 0-10000 basis points
* `expiresAt` is required and must be a future timestamp (milliseconds)
* OCO/OTOCO: take-profit price must be greater than stop-loss price
* Trailing: set exactly one of `triggerPriceUsd` or `trailingBps`; `trailingBps` must be 50–9000; `triggerMint` must be `inputMint` for `below` or `outputMint` for `above`

> [!NOTE]
>
> Unlike V1 where orders could exist indefinitely, V2 requires an expiration on every order. Adding a TTL allows the system to prune orders that are unlikely to fill, improving execution quality for active orders. For long-lived orders, use a far-future timestamp (e.g. 30 days).

## Related

> [card] [Vault & Deposit](https://developers.jup.ag/docs/trigger/deposit.md)
>
> Resolve your vault and craft the deposit this call consumes.

> [card] [Order History](https://developers.jup.ag/docs/trigger/order-history.md)
>
> Track order state, fills, and events.

> [card] [Manage Orders](https://developers.jup.ag/docs/trigger/manage-orders.md)
>
> Update trigger prices or cancel and withdraw.

> [card] [Create Order (API)](https://developers.jup.ag/docs/api-reference/trigger/create-order.md)
>
> Full create request and response schema.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.