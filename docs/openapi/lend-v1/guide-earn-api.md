> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Earn API (Beta)

> Deposit and withdraw assets to earn yield through the Jupiter Lend protocol.

> [!TIP]
>
> **API REFERENCE**
>
> To use the full Lend API, see the [Lend API Reference](https://developers.jup.ag/docs/api-reference/lend/earn/deposit.md). To choose between the API and the SDK, see [API vs SDK](https://developers.jup.ag/docs/lend/api-vs-sdk.md).

## Prerequisite

> [accordion] Dependencies
>
> ```bash theme={null}
> npm install @solana/web3.js@1 # Using v1 of web3.js instead of v2
> npm install dotenv # If required for wallet setup
> ```

> [accordion] RPC
>
> **Set up RPC**
>
> > [!NOTE]
> >
> > **NOTE**
> >
> > Solana provides a [default RPC endpoint](https://solana.com/docs/core/clusters). However, as your application grows, we recommend you to always use your own or provision a 3rd party provider’s RPC endpoint such as [Helius](https://helius.dev/) or [Triton](https://triton.one/).
>
> ```js theme={null}
> import { Connection } from "@solana/web3.js";
>
> const connection = new Connection('https://api.mainnet-beta.solana.com');
> ```

> [accordion] Wallet
>
> **Set up Development Wallet**
>
> > [!NOTE]
> >
> > **NOTE**
> >
> > * You can paste in your private key for testing purposes but this is not recommended for production applications.
> > * If you want to store your private key in the project directly, you can do it via a `.env` file.
> >
> > To set up a development wallet via `.env` file, you can use the following script.
>
> ```js theme={null}
> // index.js
> import { Keypair } from '@solana/web3.js';
> import dotenv from 'dotenv';
> require('dotenv').config();
>
> const wallet = Keypair.fromSecretKey(bs58.decode(process.env.PRIVATE_KEY || ''));
> ```
>
> ```bash theme={null}
> # .env
> PRIVATE_KEY=""
> ```
>
> To set up a development wallet via a wallet generated via [Solana CLI](https://solana.com/docs/intro/installation#solana-cli-basics), you can use the following script.
>
> ```js theme={null}
> import { Keypair } from '@solana/web3.js';
> import fs from 'fs';
>
> const privateKeyArray = JSON.parse(fs.readFileSync('/Path/To/.config/solana/id.json', 'utf8').trim());
> const wallet = Keypair.fromSecretKey(new Uint8Array(privateKeyArray));
> ```

> [accordion] Transaction Sending Example
>
> ```js theme={null}
> transaction.sign([wallet]);
> const transactionBinary = transaction.serialize();
> console.log(transactionBinary);
> console.log(transactionBinary.length);
> const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: "finalized" });
>
> const signature = await connection.sendRawTransaction(transactionBinary, {
>   maxRetries: 0,
>   skipPreflight: true,
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
> };
> ```

## Deposit and Withdraw

Using the Deposit or Withdraw endpoint, the user can do so based on the `amount` of assets to be deposited/withdrawn.

> [!NOTE]
>
> **USAGE STEPS**
>
> **Step 1**
>
> User chooses the token.
>
> **Step 2**
>
> User chooses the amount of assets to deposit or withdraw in the specific token mint.
>
> **Step 3**
>
> Post request to get the transaction.
>
> **Step 4**
>
> User sign and send the transaction to the network.
>
> **Step 5**
>
> The mint authority mints/burns the vault tokens to/from the user.

```js theme={null}
const depositTransactionResponse = await (
    await (
        await fetch('https://api.jup.ag/lend/v1/earn/deposit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': 'your-api-key',
            },
            body: JSON.stringify({
                asset: mint,
                amount: '100000',
                signer: wallet.publicKey,
            })
        })
    )
);
```

```js theme={null}
const withdrawTransactionResponse = await (
    await (
        await fetch('https://api.jup.ag/lend/v1/earn/withdraw', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': 'your-api-key',
            },
            body: JSON.stringify({
                asset: mint,
                amount: '100000',
                signer: wallet.publicKey,
            })
        })
    )
);
```

## Mint and Redeem

Using the Mint or Redeem endpoint, the user can do so based on the number `shares` to be minted/redeemed.

> [!NOTE]
>
> **USAGE STEPS**
>
> **Step 1**
>
> User chooses the token.
>
> **Step 2**
>
> User chooses the number of shares to deposit or withdraw in the specific token mint.
>
> **Step 3**
>
> Post request to get the transaction.
>
> **Step 4**
>
> User sign and send the transaction to the network.
>
> **Step 5**
>
> The mint authority mints/burns the vault tokens to/from the user.

```js theme={null}
const mintTransactionResponse = await (
    await (
        await fetch('https://api.jup.ag/lend/v1/earn/mint', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': 'your-api-key',
            },
            body: JSON.stringify({
                asset: mint,
                signer: wallet.publicKey,
                shares: '100000',
            })
        })
    )
);
```

```js theme={null}
const redeemTransactionResponse = await (
    await (
        await fetch('https://api.jup.ag/lend/v1/earn/redeem', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': 'your-api-key',
            },
            body: JSON.stringify({
                asset: mint,
                signer: wallet.publicKey,
                shares: '100000',
            })
        })
    )
);
```

## Build Your Own Transaction

The Lend API provides 2 ways to interface with the Earn functions in the Jupiter Lend Program. You can either make a post request to directly get the **Transaction**, or **Instruction** which can be used for CPI or composing with additional instructions.

### Transaction

To use the Transaction method, simply request to the endpoints without `-instructions` suffix directly, as shown in the examples above. The API will respond with an unsigned base64 transaction for the signer to sign, then sent to the network for execution.

### Instruction

In some use cases, you'd prefer to utilize the instructions instead of the serialized transaction, so you can utilize with CPI or compose with other instructions. You can make a post request to `-instructions`endpoints instead.

> [accordion] Building with Instuctions Code Snippet
>
> Example code snippet of using `/deposit-instructions` endpoint and building a transaction with the instructions.
>
> ```js theme={null}
> import { Connection, Keypair, PublicKey, TransactionMessage, TransactionInstruction, VersionedTransaction } from '@solana/web3.js';
> import fs from 'fs';
>
> const privateKeyArray = JSON.parse(fs.readFileSync('/Path/to/private/key', 'utf8').trim());
> const wallet = Keypair.fromSecretKey(new Uint8Array(privateKeyArray));
> const connection = new Connection('insert-your-own-rpc');
>
> const depositIx = await (
>     await fetch (
>         'https://api.jup.ag/lend/v1/earn/deposit-instructions', {
>             method: 'POST',
>             headers: {
>                 'Content-Type': 'application/json',
>                 'x-api-key': 'your-api-key',
>             },
>             body: JSON.stringify({
>                 asset: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
>                 amount: '1000000',
>                 signer: wallet.publicKey,
>             }, null, 2)
>         }
>     )
> ).json();
>
> console.log(JSON.stringify(depositIx, null, 2));
>
> const deserializeInstruction = (instruction) => {
>     return new TransactionInstruction({
>     programId: new PublicKey(instruction.programId),
>     keys: instruction.accounts.map((key) => ({
>         pubkey: new PublicKey(key.pubkey),
>         isSigner: key.isSigner,
>         isWritable: key.isWritable,
>     })),
>     data: Buffer.from(instruction.data, 'base64'),
>     });
> };
>
> const blockhash = (await connection.getLatestBlockhash()).blockhash;
> const messageV0 = new TransactionMessage({
>     payerKey: wallet.publicKey,
>     recentBlockhash: blockhash,
>     instructions: [
>         ...depositIx.instructions.map(deserializeInstruction)
>     ],
> }).compileToV0Message();
>
> const transaction = new VersionedTransaction(messageV0);
> transaction.sign([wallet]);
> const transactionBinary = transaction.serialize();
> console.log(transactionBinary);
> console.log(transactionBinary.length);
> const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: "finalized" });
>
> const signature = await connection.sendRawTransaction(transactionBinary, {
>   maxRetries: 0,
>   skipPreflight: true,
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
> };
> ```

### CPI

* Refer to [https://github.com/jup-ag/jupiter-lend/blob/main/docs/earn/cpi.md](https://github.com/jup-ag/jupiter-lend/blob/main/docs/earn/cpi.md) for CPI example
* Refer to [https://github.com/jup-ag/jupiter-lend/blob/main/target/idl/lending.json](https://github.com/jup-ag/jupiter-lend/blob/main/target/idl/lending.json) for IDL

## Tokens

Jupiter Lend provides Earnings for individual tokens, meaning SOL and USDC will be deposited in isolation. To get all token information such as the underlying token, supply, rates and liquidity information.

```js theme={null}
const vaults = await (
    await fetch (
        'https://api.jup.ag/lend/v1/earn/tokens',
        {
            headers: {
                'x-api-key': 'your-api-key',
            },
        }
    )
).json();
```

## User Data

Below are the endpoints to aid user to better manage their positions with data of each existing positions, earnings, etc.

### Positions

Given a user, you are able to get their existing position data such as shares, underlying assets, balance and allowance.

```js theme={null}
const userPositions = await (
    await fetch (
        'https://api.jup.ag/lend/v1/earn/positions?users={user1},{user2}'
        {
            headers: {
                'x-api-key': 'your-api-key',
            },
        }
    )
).json();
```

### Earnings

Get the earnings of specific positions for a user. The response includes the earnings amount and the slot at which it was computed.

```js theme={null}
const userEarnings = await (
    await fetch(
        'https://api.jup.ag/lend/v1/earn/earnings?user={user1}&positions={position1},{position2}',
        {
            headers: {
                'x-api-key': 'your-api-key',
            },
        }
    )
).json();
```

Response:

```json theme={null}
[
  {
    "address": "So11111111111111111111111111111111111111112",
    "ownerAddress": "BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
    "earnings": 34483000,
    "slot": 345678901
  }
]
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.