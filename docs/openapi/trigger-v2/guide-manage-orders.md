> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Orders

> Update order parameters, cancel orders with two-step withdrawal, and withdraw funds from your vault.

Update an order in place, or cancel it with a two-step withdrawal. Every request needs the JWT `token` from [authentication](https://developers.jup.ag/docs/trigger/authentication.md) and the `orderId` returned when you [created the order](https://developers.jup.ag/docs/trigger/create-order.md) (the same value as `id` in [order history](https://developers.jup.ag/docs/trigger/order-history.md)).

## Prerequisites

Signing withdrawal transactions requires `@solana/web3.js`:

```bash theme={null}
npm install @solana/web3.js
```

## Update an Order

Modify the trigger price or slippage of an existing order without cancelling and recreating it.

```typescript theme={null}
const updateResponse = await fetch(`https://api.jup.ag/trigger/v2/orders/price/${orderId}`, {
  method: 'PATCH',
  headers: {
    'Content-Type': 'application/json',
    'x-api-key': 'your-api-key',
    'Authorization': `Bearer ${token}`,
  },
  body: JSON.stringify({
    orderType: 'single',
    triggerPriceUsd: 210.00,
    slippageBps: 150,
  }),
});

const result = await updateResponse.json();
// { id: "order-uuid" }
```

### Update Fields by Order Type

**Single:**

```json theme={null}
{
  "orderType": "single",
  "triggerPriceUsd": 210.00,
  "slippageBps": 150
}
```

**OCO:**

```json theme={null}
{
  "orderType": "oco",
  "tpPriceUsd": 260.00,
  "slPriceUsd": 145.00
}
```

**OTOCO:**

```json theme={null}
{
  "orderType": "otoco",
  "triggerPriceUsd": 185.00,
  "tpPriceUsd": 230.00,
  "slPriceUsd": 155.00
}
```

**Trailing stop loss:**

```json theme={null}
{
  "orderType": "single",
  "trailingBps": 800,
  "slippageBps": 150
}
```

> [!NOTE]
>
> On a [trailing stop loss](https://developers.jup.ag/docs/trigger/create-order.md#trailing-stop-loss), update `trailingBps` (50–9000) to change the trail distance. You cannot set `triggerPriceUsd` on a trailing order (`400 "Cannot set trigger price on trailing stop loss order"`), convert a static order to trailing or back (`400 "Cannot convert to trailing stop loss order"`), or set a `trailingBps` whose new band would fire immediately (`400 "Trailing order would immediately trigger"`).

## Cancel an Order

Cancellation is a two-step process. The first step returns a withdrawal transaction that moves funds from the vault back to your wallet. You sign and submit it in the second step.

### Step 1: Initiate Cancellation

This immediately moves the order from `open` to `ready_to_cancel`. The order will no longer be filled, even if step 2 has not yet completed. This prevents a race condition where the order could still execute while you are signing the withdrawal transaction.

```typescript theme={null}
const cancelResponse = await fetch(
  `https://api.jup.ag/trigger/v2/orders/price/cancel/${orderId}`,
  {
    method: 'POST',
    headers: {
      'x-api-key': 'your-api-key',
      'Authorization': `Bearer ${token}`,
    },
  }
);

const cancelData = await cancelResponse.json();
```

**Response:**

```json theme={null}
{
  "id": "order-uuid",
  "transaction": "Base64EncodedUnsignedWithdrawalTransaction...",
  "requestId": "cancel-request-uuid"
}
```

### Step 2: Sign and Confirm

Sign the withdrawal transaction and submit it to complete the cancellation.

```typescript theme={null}
import { VersionedTransaction } from '@solana/web3.js';

const transaction = VersionedTransaction.deserialize(
  Buffer.from(cancelData.transaction, 'base64')
);
const signedTransaction = await wallet.signTransaction(transaction);

const confirmResponse = await fetch(
  `https://api.jup.ag/trigger/v2/orders/price/confirm-cancel/${orderId}`,
  {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': 'your-api-key',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      signedTransaction: Buffer.from(signedTransaction.serialize()).toString('base64'),
      cancelRequestId: cancelData.requestId,
    }),
  }
);

const result = await confirmResponse.json();
// { id: "order-uuid", txSignature: "..." }
```

If step 2 fails (the transaction doesn't land), you can retry by calling the confirm endpoint again with the same `cancelRequestId`. The order remains in `ready_to_cancel` state until the withdrawal confirms.

## Expired Order Withdrawal

If an order expires before execution, the funds remain in the vault. To retrieve them, use the same two-step cancel flow: initiate cancellation on the expired order, sign the withdrawal transaction, and confirm. The order transitions through `ready_to_cancel` and then to `cancelled` once the withdrawal confirms on-chain.

## Error Handling

| Status | Meaning |
| :- | :- |
| `400` | Invalid order ID, order not in a cancellable state, or validation error |
| `401` | Invalid API key, or JWT expired/invalid |
| `403` | Order belongs to a different wallet |
| `404` | Order not found |


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.