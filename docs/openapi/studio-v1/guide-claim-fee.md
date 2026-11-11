> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Claim Fee (Beta)

> Check and claim unclaimed LP fees from Dynamic Bonding Curve pools.

This guide covers how to claim LP fees from Dynamic Bonding Curve (DBC) pools created via Jupiter Studio. You will look up pool addresses by token mint, query total and unclaimed fees, then create and submit a claim transaction.

> [!TIP]
>
> **API REFERENCE**
>
> To fully utilize the Studio API, check out the [Studio API Reference](https://developers.jup.ag/docs/api-reference/studio).

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
> import { Connection } from '@solana/web3.js';
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
>
> To set up a development wallet via `.env` file, you can use the following script.
>
> ```js theme={null}
> // index.js
> import { Keypair } from '@solana/web3.js';
> import dotenv from 'dotenv';
> require('dotenv').config();
> const wallet = Keypair.fromSecretKey(bs58.decode(process.env.PRIVATE_KEY || ''));
> ```
>
> ```bash theme={null}
> # .envPRIVATE_KEY=''
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
> const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: 'finalized' });
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
>   }, 'confirmed');
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

## Pool Address

Your successfully created token via Jupiter Studio, should have a newly generated token mint. By using the mint, you can get the config key and pool addresses associated to it: Dynamic Bonding Curve pool and Meteora DAMM V2 pool.

```js theme={null}
const poolAddressResponse = await (
    await fetch(
      `https://api.jup.ag/studio/v1/dbc-pool/addresses/${mint}`,
      {
        headers: {
          'x-api-key': 'your-api-key',
        },
      }
    )
).json();
```

## Fee

Using the Pool Address, you will be able to get the total and current unclaimed fees in the Dynamic Bonding Curve pool.

```js theme={null}
const feeResponse = await (
    await fetch (
      'https://api.jup.ag/studio/v1/dbc/fee',
      {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': 'your-api-key',
        },
        body: JSON.stringify({
            poolAddress: poolAddressResponse.data.dbcPoolAddress,
        }, null, 2)
    })
).json();
```

## Claim Fee

In order to claim fees from a Dynamic Bonding Curve pool, you will need to pass in the pool address into this endpoint and we will create the Claim Fee transaction for you. After receiving the transaction, you will need to sign and submit the transaction to the network on your own ([refer to Transaction Sending Example above](#prerequisite)).

```js theme={null}
const claimTransaction = await (
    await fetch (
      'https://api.jup.ag/studio/v1/dbc/fee/create-tx',
      {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': 'your-api-key',
        },
        body: JSON.stringify({
            ownerWallet: wallet.publicKey.toBase58(),
            poolAddress: poolAddressResponse.data.dbcPoolAddress,
            maxQuoteAmount: 1000000, // e.g. 1 USDC (depending on quote mint and decimals)
        }, null, 2)
    })
).json();
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.