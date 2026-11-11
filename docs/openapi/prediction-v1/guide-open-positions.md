> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Open Positions

> Create buy orders to open YES or NO positions in Jupiter Prediction Markets.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc walks through creating a buy order to open a position in a prediction market. You'll learn how to construct the order request, sign the returned transaction, and submit it to the Solana network.

## Prerequisite

Before opening a position, ensure you have:

* An API key from the [Portal](https://developers.jup.ag/portal)
* A Solana wallet with funds (JupUSD or USDC)
* Identified a market to trade (see [Events & Markets](https://developers.jup.ag/docs/prediction/events-and-markets.md))
* Requires signing and transaction submission, [install `@solana/web3.js` npm library](https://developers.jup.ag/docs/get-started/environment-setup)

> [!TIP]
>
> **API REFERENCE**
>
> For the complete endpoint specification, see the [Create Order API Reference](https://developers.jup.ag/docs/api-reference/prediction/create-order.md).

***

## Create Buy Order

Use `POST /orders` to create a buy order for YES or NO contracts.

| Parameter | Type | Required | Description |
| :- | :- | :- | :- |
| `ownerPubkey` | string | Yes | Your wallet's public key |
| `marketId` | string | Yes | The market identifier |
| `isYes` | boolean | Yes | `true` for YES contracts, `false` for NO |
| `isBuy` | boolean | Yes | `true` for buy orders |
| `contracts` | string | No | Number of contracts to purchase |
| `depositAmount` | string | Yes | Amount to deposit (in native token units of JupUSD or USDC) |
| `depositMint` | string | Yes | Token mint for deposit (JupUSD or USDC) |

The minimum order is 5 USD (`5000000` native token units). Smaller deposits return `400 Minimum order is $5`.

Example: Create a buy order for 5 USD worth of contracts

```js theme={null}
const response = await fetch('https://api.jup.ag/prediction/v1/orders', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'x-api-key': 'your-api-key'
  },
  body: JSON.stringify({
    ownerPubkey: wallet.publicKey.toString(),
    depositAmount: '5000000',
    depositMint: 'JuprjznTrTSp2UFa3ZBUFgwdAmtZCq4MQCwysN55USD',
    marketId: marketId,
    isYes: true,
    isBuy: true,
  })
});

const orderResponse = await response.json();
console.log(orderResponse);
```

## Understanding the Response

| Field | Description |
| :- | :- |
| `transaction` | Base64-encoded Solana transaction to sign and submit |
| `txMeta` | Transaction metadata: `blockhash` and `lastValidBlockHeight` |
| `order.orderPubkey` | The order's on-chain account address |
| `order.positionPubkey` | The position's on-chain account address |
| `order.contracts` | Legacy whole-contract quantity. Contracts are fractional: use `order.contractsMicro` (`1000000` = 1 contract) or `order.contractsDecimal` for the exact amount |
| `order.requiredSigners` | Public keys that must sign the transaction |

## Sign and Submit the Transaction

After receiving the order response, you need to:

1. Deserialize the base64 transaction
2. Sign it with your wallet
3. Submit to the Solana network

> [accordion] Full Code Example
>
> ```js theme={null}
> // Step 1: Create order
> const orderResponse = await fetch('https://api.jup.ag/prediction/v1/orders', {
>   method: 'POST',
>   headers: {
>     'Content-Type': 'application/json',
>     'x-api-key': 'your-api-key'
>   },
>   body: JSON.stringify({
>     ownerPubkey: wallet.publicKey.toString(),
>     depositAmount: '5000000',
>     depositMint: 'JuprjznTrTSp2UFa3ZBUFgwdAmtZCq4MQCwysN55USD',
>     marketId: marketId, // Get from events and markets guide
>     isYes: true,
>     isBuy: true,
>   })
> }).then(r => r.json());
>
> // Step 2: Deserialize and sign transaction
> const transaction = VersionedTransaction.deserialize(
>   Buffer.from(orderResponse.transaction, 'base64')
> );
> transaction.sign([wallet]);
>
> const transactionBinary = transaction.serialize();
> const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: "confirmed" });
>
> // Step 3: Send the transaction and await status
> const signature = await connection.sendRawTransaction(transactionBinary, {
>   maxRetries: 0,
>   skipPreflight: true,
>   preflightCommitment: "confirmed",
> });
>
> console.log(`Transaction sent: https://solscan.io/tx/${signature}`);
>
> try {
>   const confirmation = await connection.confirmTransaction({
>     signature,
>     blockhash: blockhashInfo.value.blockhash,
>     lastValidBlockHeight: blockhashInfo.value.lastValidBlockHeight,
>   }, "confirmed");
>
>   if (confirmation.value.err) {
>     console.error(`Transaction failed: ${JSON.stringify(confirmation.value.err)}`);
>     console.log(`Examine the failed transaction: https://solscan.io/tx/${signature}`);
>   } else {
>     console.log(`Transaction successful: https://solscan.io/tx/${signature}`);
>   }
> } catch (error) {
>   console.error(`Error confirming transaction: ${error}`);
>   console.log(`Examine the transaction status: https://solscan.io/tx/${signature}`);
> }
> ```

## Checking Order Status

After submitting, the order enters the keeper network for matching and execution. Use `GET /orders/status/{orderPubkey}` to check the fill status.

> [!TIP]
>
> **Order Flow**
>
> 1. **Create Order**: You sign a transaction that creates an order account on-chain - this transaction only opens an order account and **does not** guarantee that the order will be filled.
> 2. **Keeper Fills**: Jupiter's keeper network matches your order with the underlying prediction market
> 3. **Position Updated**: Once filled, your position account reflects the new contracts
> 4. **Order Closed**: The order account is closed after completion

> [!TIP]
>
> Polling immediately after submitting the transaction might return a `pending` status or a `no order history found` error. Wait for a few slots before polling again.

| Status | Description |
| :- | :- |
| `created` | Order account created on-chain, waiting to be filled |
| `partiallyfilled` | Order partially filled |
| `filled` | Order fully executed |
| `failed` | Order could not be filled |

`GET /orders/{orderPubkey}` returns `400` once an order has filled and its account is closed; use `GET /orders/status/{orderPubkey}` or `GET /history` to track a completed order.

```js theme={null}
const orderPubkey = orderResponse.order.orderPubkey;

const statusResponse = await fetch(
  `https://api.jup.ag/prediction/v1/orders/status/${orderPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const status = await statusResponse.json();
console.log(status);
```

***

## Fee Calculation

To understand how fees work, refer to the [Fees](https://developers.jup.ag/docs/prediction.md#fees) page.

***

## What's Next

Once you have open positions, learn how to manage them - view your holdings, sell contracts, or cancel pending orders.

> [card] [Manage Positions](https://developers.jup.ag/docs/prediction/manage-positions.md)
>
> View, sell, and manage your positions

> [card] [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md)
>
> Query positions, orders, and transaction history


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.