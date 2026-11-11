> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Positions

> View your prediction market positions and sell contracts to close positions.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc covers how to view your open positions, sell contracts to reduce or exit positions, and cancel pending orders.

## Prerequisite

Before managing positions, ensure you have:

* An API key from the [Portal](https://developers.jup.ag/portal)
* Open positions from [Opening Positions](https://developers.jup.ag/docs/prediction/open-positions.md)
* Requires signing and transaction submission, [install `@solana/web3.js` npm library](https://developers.jup.ag/docs/get-started/environment-setup)

> [!TIP]
>
> **API REFERENCE**
>
> For complete endpoint specifications, see the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction).

***

## Viewing Your Positions

See [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md) for full details on how to query, filter, and interpret your open positions, including all available parameters and data fields.

Example: Fetch all your open positions

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
console.log(positions);
```

> [!NOTE]
>
> For descriptions of all position fields, parameters, and how to aggregate P\&L,
> see the [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md) doc.
>
> All USD values returned are denominated in native token units of JupUSD or USDC, where **1,000,000 units = \$1.00**.

***

## Viewing Your Orders

See [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md#querying-orders) for details on available parameters, order fields, and examples.

Example: Fetch all your open orders

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

***

## Selling Contracts

You can sell contracts by closing entire position.

### Close Entire Position

Use `DELETE /positions/{positionPubkey}` to sell all contracts in a position.

```js theme={null}
const positionPubkey = 'position-pubkey-789';

const response = await fetch(
  `https://api.jup.ag/prediction/v1/positions/${positionPubkey}`,
  {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': 'your-api-key'
    },
    body: JSON.stringify({
      ownerPubkey: wallet.publicKey.toString(),
    })
  }
);

const closeResponse = await response.json();
// Sign and submit the transaction (see below)
```

### Close All Positions

Use `DELETE /positions` to close all your open positions at once. This is useful for quickly exiting all markets.

`minSellPriceSlippageBps` is required (slippage tolerance per sell, in basis points). The request returns `{ data: [...] }`, one sell transaction per open position, each in the same shape as a single close.

```js theme={null}
const response = await fetch('https://api.jup.ag/prediction/v1/positions', {
  method: 'DELETE',
  headers: {
    'Content-Type': 'application/json',
    'x-api-key': 'your-api-key'
  },
  body: JSON.stringify({
    ownerPubkey: wallet.publicKey.toString(),
    minSellPriceSlippageBps: 2500,
  })
});

const closeAllResponse = await response.json();
// Sign and submit each transaction in closeAllResponse.data (see above)
```

### Sign and Submit the Transaction

You'll receive a base64-encoded transaction. You need to sign and submit it to finalize the position closure on-chain.

> [accordion] Full Close Position Code Example
>
> ```js theme={null}
> import {
>   Connection,
>   Keypair,
>   VersionedTransaction,
> } from '@solana/web3.js';
> import fs from 'fs';
>
> // Setup connection and wallet
> const connection = new Connection('YOUR_RPC_URL');
> const privateKey = JSON.parse(fs.readFileSync('/path/to/wallet.json', 'utf8'));
> const wallet = Keypair.fromSecretKey(new Uint8Array(privateKey));
>
> // Step 1: Initiate closing a position
> const positionPubkey = 'YOUR_POSITION_PUBKEY';
>
> const response = await fetch(
>   `https://api.jup.ag/prediction/v1/positions/${positionPubkey}`,
>   {
>     method: 'DELETE',
>     headers: {
>       'Content-Type': 'application/json',
>       'x-api-key': 'your-api-key'
>     },
>     body: JSON.stringify({
>       ownerPubkey: wallet.publicKey.toString(),
>     })
>   }
> );
>
> const closeResponse = await response.json();
> const base64Txn = closeResponse.transaction;
>
> // Step 2: Deserialize and sign transaction
> const transaction = VersionedTransaction.deserialize(
>   Buffer.from(base64Txn, 'base64')
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
>     console.log(`Position closed successfully: https://solscan.io/tx/${signature}`);
>   }
> } catch (error) {
>   console.error(`Error confirming transaction: ${error}`);
>   console.log(`Examine the transaction status: https://solscan.io/tx/${signature}`);
> }
> ```

***

## What's Next

Learn how to claim payouts when markets settle in your favor, or explore the data endpoints for tracking your activity.

> [card] [Claim Payouts](https://developers.jup.ag/docs/prediction/claim-payouts.md)
>
> Claim winnings after market settlement

> [card] [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md)
>
> Query positions, orders, and transaction history


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.