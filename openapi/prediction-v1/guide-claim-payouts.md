> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Claim Payouts

> Claim your winnings after prediction markets resolve. Each winning contract pays out $1.

> [!NOTE]
>
> **BETA**
>
> The Prediction Market API is currently in beta and subject to breaking changes as we continue to improve the product. If you have any feedback, please reach out in [Discord](https://discord.gg/jup).

This doc explains how to claim payouts after a prediction market resolves in your favor. Each winning contract is worth \$1.00, and there are no fees for claiming.

## Prerequisite

Before claiming payouts, ensure you have:

* An API key from the [Portal](https://developers.jup.ag/portal)
* [Opened positions](https://developers.jup.ag/docs/prediction/open-positions.md) and have settled in your favor
* Requires signing and transaction submission, [install `@solana/web3.js` npm library](https://developers.jup.ag/docs/get-started/environment-setup)

> [!TIP]
>
> **API REFERENCE**
>
> For complete endpoint specifications, see the [Prediction API Reference](https://developers.jup.ag/docs/api-reference/prediction).

***

## Understanding Market Settlement

When a market closes, it goes through a settlement process:

### Market States

| State | Description |
| :- | :- |
| `open` | Market actively trading |
| `closed` | Trading stopped; awaiting or holding the result |
| `cancelled` | Market voided, positions refunded |

A resolved market stays `closed` with its `result` set; there is no separate `settled` status.

### Market Results

| Result | Description |
| :- | :- |
| `null` | Not yet resolved |
| `yes` | YES outcome confirmed |
| `no` | NO outcome confirmed |

### Position States After Settlement

| State | Description |
| :- | :- |
| **Claimable** | Market settled in your favor (`claimable: true`) |
| **Lost** | Market settled against your position |
| **Claimed** | Payout already withdrawn (`claimed: true`) |

> [!NOTE]
>
> **Winning Payout**
>
> Each winning contract pays out exactly **\$1.00** (1,000,000 native token units for JupUSD or USDC). If you hold 10 YES contracts and the market resolves to YES, your payout is \$10.00.
>
> There are **no fees** for claiming payouts.

> [!NOTE]
>
> If you do not manually claim a payout from a winning position within 24 hours of the market settling in your favor, your claimable position will be **automatically claimed on your behalf** by the keeper network. The payout will be credited to your account wallet, and the position will be marked as claimed.
>
> * You can always claim manually as soon as your position is claimable.
> * Auto-claim ensures you never lose out on winnings if you forget to claim.
> * The payout amount is the same as a manual claim, and no fees are charged for auto-claiming.
> * You can check your claim status after settlement and see the updated `claimed: true` field and your payout even if you did not claim manually.

***

## Checking for Claimable Positions

Query your positions and filter for those that are claimable.

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

// Filter for claimable positions
const claimablePositions = positions.data.filter(position =>
  position.claimable === true && position.claimed === false
);

console.log('Claimable positions:', claimablePositions.length);

for (const position of claimablePositions) {
  const payoutUsd = parseInt(position.payoutUsd) / 1_000_000;
  console.log(`Position ${position.pubkey}: $${payoutUsd} payout available`);
}
```

Example claimable position:

```json expandable theme={null}
{
  "pubkey": "position-pubkey-789",
  "ownerPubkey": "your-wallet-pubkey",
  "marketId": "market-456",
  "isYes": true,
  "contracts": "10",
  "totalCostUsd": "6500000",
  "avgPriceUsd": "650000",
  "payoutUsd": "10000000",
  "claimable": true,
  "claimed": false,
  "claimedUsd": "0",
  "marketMetadata": {
    "title": "SOL $500 by Q2 2026",
    "status": "closed",
    "result": "yes"
  }
}
```

> [!TIP]
>
> **Checking Eligibility**
>
> A position is claimable when:
>
> * `claimable: true` - Market settled in your favor
> * `claimed: false` - Payout hasn't been withdrawn yet
> * `marketMetadata.result` matches your position side (e.g., `yes` for YES positions)

## Claiming a Payout

Use `POST /positions/{positionPubkey}/claim` to create a claim transaction. Calling it before the market resolves returns `400 Market not settled yet`.

> [!NOTE]
>
> Some markets settle automatically and do not need a claim. For those positions the endpoint returns a message saying no claim is required. Combined with the 24-hour auto-claim, you can always rely on winnings being credited without a manual claim.

```js theme={null}
const positionPubkey = 'position-pubkey-789';

const response = await fetch(
  `https://api.jup.ag/prediction/v1/positions/${positionPubkey}/claim`,
  {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': 'your-api-key'
    },
    body: JSON.stringify({
      ownerPubkey: wallet.publicKey.toString()
    })
  }
);

const claimResponse = await response.json();
console.log(claimResponse);
```

Example response:

```json expandable theme={null}
{
  "transaction": "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAQAGDkS+3LuGTbs...",
  "txMeta": {
    "blockhash": "EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N",
    "lastValidBlockHeight": 279632475
  },
  "position": {
    "positionPubkey": "position-pubkey-789",
    "marketPubkey": "market-pubkey-456",
    "userPubkey": "your-wallet-pubkey",
    "isYes": true,
    "contracts": "10",
    "contractsMicro": "10000000",
    "contractsDecimal": "10",
    "payoutAmountUsd": "10000000"
  }
}
```

## Sign and Submit Claim Transaction

After receiving the claim response, sign and submit the transaction.

> [accordion] Full Claim Code Example
>
> ```js theme={null}
> import {
>   Connection,
>   Keypair,
>   VersionedTransaction,
> } from '@solana/web3.js';
> import fs from 'fs';
>
> // Setup
> const connection = new Connection('YOUR_RPC_URL');
> const privateKey = JSON.parse(fs.readFileSync('/path/to/wallet.json', 'utf8'));
> const wallet = Keypair.fromSecretKey(new Uint8Array(privateKey));
>
> // Step 1: Find claimable positions
> const positionsResponse = await fetch(
>   'https://api.jup.ag/prediction/v1/positions?' + new URLSearchParams({
>     ownerPubkey: wallet.publicKey.toString()
>   }),
>   {
>     headers: {
>       'x-api-key': 'your-api-key'
>     }
>   }
> ).then(r => r.json());
>
> const claimable = positionsResponse.data.filter(p =>
>   p.claimable === true && p.claimed === false
> );
>
> if (claimable.length === 0) {
>   console.log('No claimable positions found');
>   process.exit(0);
> }
>
> // Step 2: Claim each position
> for (const position of claimable) {
>   console.log(`Claiming position ${position.pubkey}...`);
>
>   const claimResponse = await fetch(
>     `https://api.jup.ag/prediction/v1/positions/${position.pubkey}/claim`,
>     {
>       method: 'POST',
>       headers: {
>         'Content-Type': 'application/json',
>         'x-api-key': 'your-api-key'
>       },
>       body: JSON.stringify({
>         ownerPubkey: wallet.publicKey.toString()
>       })
>     }
>   ).then(r => r.json());
>
>   // Step 3: Deserialize and sign transaction
>   const transaction = VersionedTransaction.deserialize(
>     Buffer.from(claimResponse.transaction, 'base64')
>   );
>   transaction.sign([wallet]);
>
>   const transactionBinary = transaction.serialize();
>   const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: "confirmed" });
>
>   // Step 4: Send the transaction and await status
>   const signature = await connection.sendRawTransaction(transactionBinary, {
>     maxRetries: 0,
>     skipPreflight: true,
>     preflightCommitment: "confirmed",
>   });
>
>   console.log(`Transaction sent: https://solscan.io/tx/${signature}`);
>
>   try {
>     const confirmation = await connection.confirmTransaction({
>       signature,
>       blockhash: blockhashInfo.value.blockhash,
>       lastValidBlockHeight: blockhashInfo.value.lastValidBlockHeight,
>     }, "confirmed");
>
>     if (confirmation.value.err) {
>       console.error(`Transaction failed: ${JSON.stringify(confirmation.value.err)}`);
>       console.log(`Examine the failed transaction: https://solscan.io/tx/${signature}`);
>     } else {
>       const payoutUsd = parseInt(position.payoutUsd) / 1_000_000;
>       console.log(`Transaction successful: https://solscan.io/tx/${signature}`);
>       console.log(`Successfully claimed $${payoutUsd}!`);
>     }
>   } catch (error) {
>     console.error(`Error confirming transaction: ${error}`);
>     console.log(`Examine the transaction status: https://solscan.io/tx/${signature}`);
>   }
> }
> ```

## Verifying the Claim

After claiming, you can verify by fetching the position again.

```js theme={null}
const response = await fetch(
  `https://api.jup.ag/prediction/v1/positions/${positionPubkey}`,
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const position = await response.json();
console.log('Claimed:', position.claimed); // Should be true
console.log('Claimed amount:', position.claimedUsd);
```

## Handling Lost Positions

When a market settles against your position, the contracts expire worthless. No action is needed - the position simply has no claimable value.

Lost positions appear in your transaction history with the `position_lost` event type.

```js theme={null}
// Check history for lost positions
const historyResponse = await fetch(
  'https://api.jup.ag/prediction/v1/history?' + new URLSearchParams({
    ownerPubkey: wallet.publicKey.toString()
  }),
  {
    headers: {
      'x-api-key': 'your-api-key'
    }
  }
);

const history = await historyResponse.json();
const lostPositions = history.data.filter(event =>
  event.eventType === 'position_lost'
);
```

## Summary

| Scenario | Action |
| :- | :- |
| Position is claimable | Call `POST /positions/{pubkey}/claim` or wait for auto-claim (happens automatically after 24 hours if you don't claim manually) |
| Position already claimed | No action needed (`claimed: true`) |
| Position lost | No action needed (contracts worthless) |
| Market cancelled | Position refunded automatically |

***

## What's Next

Track your complete trading activity including claims, P\&L, and transaction history.

> [card] [Position Data & History](https://developers.jup.ag/docs/prediction/position-data.md)
>
> Query positions, orders, and transaction history

> [card] [Social Features](https://developers.jup.ag/docs/prediction/social-features.md)
>
> View profiles, leaderboards, and trade feeds


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.