> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Express Verification API

> Programmatically submit tokens for verification and update metadata via API

The Express Verification API lets you submit tokens for [Jupiter VRFD](https://verified.jup.ag) verification and update token metadata programmatically. This is the same Express flow available on the VRFD UI, exposed as an API for launchpads, agents, and projects that need to integrate verification into their token creation pipelines.

Each Express submission costs 1000 JUP, which prevents spam and prioritizes requests. You can pay in JUP directly, or in SOL, USDC, or JUPUSD, which are converted to 1000 JUP via Ultra. Standard submissions on the [VRFD site](https://verified.jup.ag) are free.

You can request verification and metadata updates together. Each is reviewed independently, so a metadata update can proceed even if verification is declined.

## Payment Currencies

Pass `paymentCurrency` as a token symbol (for example `SOL`), not a mint address. JUP is the default and transfers directly. SOL, USDC, and JUPUSD are converted to 1000 JUP via an Ultra swap.

| `paymentCurrency` value | Resolves to mint | Decimals | Conversion |
| - | - | - | - |
| `JUP` | `JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN` | 6 | Direct transfer of 1000 JUP |
| `SOL` | `So11111111111111111111111111111111111111112` | 9 | Swapped to 1000 JUP via Ultra |
| `USDC` | `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v` | 6 | Swapped to 1000 JUP via Ultra |
| `JUPUSD` | `JuprjznTrTSp2UFa3ZBUFgwdAmtZCq4MQCwysN55USD` | 6 | Swapped to 1000 JUP via Ultra |

## Prerequisites

1. Get an API key at [Portal](https://developers.jup.ag/portal). All requests require the `x-api-key` header.
2. The submitting wallet needs enough of your chosen payment currency for 1000 JUP. Transactions are gasless, so the wallet does not need extra SOL for network fees.

## Flow

The API is a three-step flow: check eligibility, craft a payment transaction, sign it, then execute.

### Step 1: Check Eligibility (Optional)

Check if the token can be submitted for verification and/or metadata updates. This step is optional since the execute step also rejects ineligible tokens before charging payment, but it keeps your flow clean and avoids unnecessary transaction crafting.

The most common reason for ineligibility is an existing pending submission for the token.

```typescript theme={null}
const response = await fetch(
  'https://api.jup.ag/tokens/v2/verify/express/check-eligibility?tokenId=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
  {
    headers: { 'x-api-key': 'YOUR_API_KEY' },
  }
);
const eligibility = await response.json();
```

```json theme={null}
{
  "tokenExists": true,
  "isVerified": false,
  "canVerify": true,
  "canMetadata": false,
  "metadataError": "A premium metadata update request for this token already exists"
}
```

| Field | Type | Description |
| - | - | - |
| `tokenExists` | `boolean` | Whether the token exists on-chain |
| `isVerified` | `boolean` | Whether the token is already verified |
| `canVerify` | `boolean` | Whether a verification request can be submitted |
| `canMetadata` | `boolean` | Whether a metadata update can be submitted |
| `verificationError` | `string` | Why verification cannot proceed (only present when there is a specific error) |
| `metadataError` | `string` | Why metadata cannot proceed (only present when there is a specific error) |

If both `canVerify` and `canMetadata` are `false`, the submission will be rejected before any payment is charged.

### Step 2: Craft the Payment Transaction

This returns an unsigned Solana transaction worth 1000 JUP. By default it crafts a direct JUP transfer. Pass `paymentCurrency` to pay in SOL, USDC, or JUPUSD instead, which crafts an Ultra swap to JUP (see [Pay with SOL, USDC, or JUPUSD](#pay-with-sol-usdc-or-jupusd)). Save the `requestId` from the response, you need it in the next step.

```typescript theme={null}
const craftResponse = await fetch(
  `https://api.jup.ag/tokens/v2/verify/express/craft-txn?senderAddress=${walletAddress}`,
  {
    headers: { 'x-api-key': 'YOUR_API_KEY' },
  }
);
const { transaction, requestId } = await craftResponse.json();
```

| Parameter | Type | Required | Description |
| - | - | - | - |
| `senderAddress` | `string` | Yes | Solana wallet address that will sign and pay for the transaction |
| `paymentCurrency` | `string` | No | Token symbol, not a mint address: `JUP` (default), `SOL`, `USDC`, or `JUPUSD`. Non-JUP currencies route through an Ultra swap to JUP. |

The response includes the base64-encoded unsigned transaction and a `requestId` that links this payment to your execute request.

```json theme={null}
{
  "receiverAddress": "5x...recipient",
  "mint": "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
  "amount": "1000000000",
  "tokenDecimals": 6,
  "tokenUsdRate": 0.42,
  "feeLamports": 5000,
  "feeUsdAmount": 0.0009,
  "feeMint": "So11111111111111111111111111111111111111112",
  "feeTokenDecimals": 9,
  "feeAmount": 5000,
  "transaction": "AQAB...",
  "requestId": "req_abc123",
  "totalTime": 412,
  "expireAt": "2026-06-30T12:00:00.000Z",
  "code": 0,
  "gasless": true
}
```

| Field | Type | Description |
| - | - | - |
| `transaction` | `string` | Base64-encoded unsigned transaction |
| `requestId` | `string` | Unique ID for this payment. Pass to `/execute`. |
| `receiverAddress` | `string` | Wallet address receiving the JUP payment |
| `mint` | `string` | Mint of the JUP collected (always JUP) |
| `amount` | `string` | JUP amount in smallest unit (1000 JUP = `1000000000`) |
| `tokenDecimals` | `number` | Decimals of the JUP collected (6) |
| `tokenUsdRate` | `number` | USD rate of JUP at craft time |
| `feeLamports` | `number` | Network fee for the transaction, in lamports |
| `feeUsdAmount` | `number` | Network fee in USD |
| `feeMint` | `string` | Mint address of the fee token |
| `feeTokenDecimals` | `number` | Decimals of the fee token |
| `feeAmount` | `number` | Fee amount in the fee token's smallest unit |
| `totalTime` | `number` | Time taken to craft the transaction, in milliseconds |
| `expireAt` | `string` | Expiry timestamp for the crafted transaction |
| `code` | `number` | Status code (`0` for success) |
| `gasless` | `boolean` | Whether the transaction is gasless |

Non-JUP payments return additional swap sizing fields. See [Pay with SOL, USDC, or JUPUSD](#pay-with-sol-usdc-or-jupusd).

### Step 3: Sign and Execute

Sign the transaction from Step 2 with your wallet, then submit it along with the token details and metadata.

```typescript theme={null}
import { VersionedTransaction } from '@solana/web3.js';

// Deserialize and sign the transaction
const txBuffer = Buffer.from(transaction, 'base64');
const tx = VersionedTransaction.deserialize(txBuffer);
tx.sign([wallet]); // wallet is a Keypair

// Serialize the signed transaction back to base64
const signedTransaction = Buffer.from(tx.serialize()).toString('base64');

// Submit
const executeResponse = await fetch(
  'https://api.jup.ag/tokens/v2/verify/express/execute',
  {
    method: 'POST',
    headers: {
      'x-api-key': 'YOUR_API_KEY',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      transaction: signedTransaction,
      requestId,
      senderAddress: wallet.publicKey.toBase58(),
      tokenId: 'YOUR_TOKEN_MINT_ADDRESS',
      twitterHandle: 'https://x.com/yourproject',
      description: 'Submitting for verification because...',
      tokenMetadata: {
        tokenId: 'YOUR_TOKEN_MINT_ADDRESS',
        website: 'https://yourproject.io',
        twitter: 'https://x.com/yourproject',
        discord: 'https://discord.gg/yourproject',
      },
    }),
  }
);
const result = await executeResponse.json();
```

| Field | Type | Required | Description |
| - | - | - | - |
| `transaction` | `string` | Yes | Base64-encoded signed transaction from craft-txn |
| `requestId` | `string` | Yes | `requestId` from the craft-txn response |
| `senderAddress` | `string` | Yes | Solana wallet address of the sender |
| `tokenId` | `string` | Yes | Token mint address to verify |
| `twitterHandle` | `string` | Yes | Twitter/X handle of the token project |
| `description` | `string` | Yes | Reason for the verification request |
| `senderTwitterHandle` | `string` | No | Twitter/X handle of the submitter |
| `tokenMetadata` | `object` | No | Token metadata to submit alongside verification. See [Token Metadata Fields](#token-metadata-fields). |
| `paymentCurrency` | `string` | No | Token symbol, not a mint address: `JUP` (default), `SOL`, `USDC`, or `JUPUSD`. Must match the currency used in craft-txn. |
| `paymentAmount` | `string` | Conditional | Atomic input amount paid, from the craft response's `quotedInputAmount`. Required when `paymentCurrency` is not `JUP`. |
| `jupOutputAmount` | `string` | No | Atomic JUP output from the craft response (its `amount` field). Optional, but recommended on non-JUP paths so revenue tracks the JUP actually collected. |

```json theme={null}
{
  "status": "Success",
  "signature": "5UfDuX...",
  "totalTime": 1234,
  "verificationCreated": true,
  "metadataCreated": true
}
```

| Field | Type | Description |
| - | - | - |
| `status` | `"Success" \| "Failed"` | Whether the transaction executed |
| `signature` | `string` | On-chain transaction signature (present on success) |
| `error` | `string` | Error message (present on failure) |
| `code` | `number` | Error code (present on failure) |
| `totalTime` | `number` | Time taken to execute, in milliseconds |
| `verificationCreated` | `boolean` | Whether a verification request was created |
| `metadataCreated` | `boolean` | Whether a metadata update request was created |

After submission, track your request status at [verified.jup.ag/tokens/browse](https://verified.jup.ag/tokens/browse).

## Pay with SOL, USDC, or JUPUSD

To pay with a non-JUP currency, set `paymentCurrency` on both the craft and execute steps. The craft step returns an Ultra swap that converts your input token into JUP, sized with a small buffer so the swap yields at least 1000 JUP.

Request the transaction with your chosen currency:

```typescript theme={null}
const craftResponse = await fetch(
  `https://api.jup.ag/tokens/v2/verify/express/craft-txn?senderAddress=${walletAddress}&paymentCurrency=SOL`,
  {
    headers: { 'x-api-key': 'YOUR_API_KEY' },
  }
);
const craft = await craftResponse.json();
```

On non-JUP paths the response includes these swap sizing fields in addition to the common fields above:

| Field | Type | Description |
| - | - | - |
| `inputMint` | `string` | Input token mint address |
| `inputDecimals` | `number` | Input token decimals |
| `quotedInputAmount` | `string` | Atomic input amount for the swap, sized to yield at least 1000 JUP |
| `maxInputAmount` | `string` | Maximum atomic input amount for the swap. Equal to `quotedInputAmount` on the current swap flow. |

When you execute, pass `paymentAmount` (required on non-JUP paths) and `jupOutputAmount` (optional, for revenue tracking). `paymentAmount` is the input amount from the craft response's `quotedInputAmount`, and `jupOutputAmount` is the craft response's `amount`:

```typescript theme={null}
body: JSON.stringify({
  transaction: signedTransaction,
  requestId,
  senderAddress: wallet.publicKey.toBase58(),
  tokenId: 'YOUR_TOKEN_MINT_ADDRESS',
  twitterHandle: 'https://x.com/yourproject',
  description: 'Submitting for verification because...',
  paymentCurrency: 'SOL',
  paymentAmount: craft.quotedInputAmount,
  jupOutputAmount: craft.amount,
}),
```

## Token Metadata Fields

The `tokenMetadata` object in the execute request is optional. Include it to update metadata alongside your verification submission. Only `tokenId` is required, all other fields are optional.

| Field | Type | Description |
| - | - | - |
| `tokenId` | `string` | Token mint address (required) |
| `name` | `string` | Token display name |
| `symbol` | `string` | Token ticker symbol |
| `icon` | `string` | URL to token icon image |
| `website` | `string` | Project website URL |
| `twitter` | `string` | Twitter/X profile URL |
| `twitterCommunity` | `string` | Twitter/X community URL |
| `telegram` | `string` | Telegram group URL |
| `discord` | `string` | Discord invite URL |
| `instagram` | `string` | Instagram profile URL |
| `tiktok` | `string` | TikTok profile URL |
| `tokenDescription` | `string` | Short description of the token |
| `circulatingSupply` | `string` | Circulating supply value |
| `useCirculatingSupply` | `boolean` | Whether to use the provided `circulatingSupply` value |
| `circulatingSupplyUrl` | `string` | URL to a live circulating supply endpoint |
| `useCirculatingSupplyUrl` | `boolean` | Whether to use the provided `circulatingSupplyUrl` |
| `coingeckoCoinId` | `string` | CoinGecko coin ID for price data |
| `useCoingeckoCoinId` | `boolean` | Whether to use the provided `coingeckoCoinId` |
| `otherUrl` | `string` | Any other relevant URL |

## Known Projects and Launchpads

If you are a known project or launchpad on Solana, DM [@jup\_vrfd](https://x.com/jup_vrfd) from your organization account on X. This links your Developer Platform OrgId to your submissions, which helps the VRFD team review them.

## API Reference

> [card] [Check Eligibility](https://developers.jup.ag/docs/api-reference/tokens/verify-check-eligibility.md)
>
> Check if a token can be verified

> [card] [Craft Transaction](https://developers.jup.ag/docs/api-reference/tokens/verify-craft-txn.md)
>
> Get an unsigned 1000 JUP payment transaction

> [card] [Execute](https://developers.jup.ag/docs/api-reference/tokens/verify-execute.md)
>
> Submit signed transaction with token details


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.