> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Token (Beta)

> Launch a token with configurable bonding curves, metadata, and images through the Studio API.

This guide walks through the full token creation flow using the Studio API: calling the `/studio/v1/dbc-pool/create-tx` endpoint to generate a transaction and presigned URLs, uploading token image and metadata, then signing and submitting via `/studio/v1/dbc-pool/submit`.

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
>
> const wallet = Keypair.fromSecretKey(bs58.decode(process.env.PRIVATE_KEY || ''));
> ```
>
> ```bash theme={null}
> # .env
> PRIVATE_KEY=''
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

## Create Transaction

This endpoint helps you create a few key components to launch your token on Studio.

1. `transaction`: A base64-encoded unsigned transaction.
2. `mint`: The mint of the token that is being created.
3. `imagePresignedUrl`: A `PUT` request endpoint to upload your token image.
4. `metadataPresignedUrl`: A `PUT` request endpoint to upload your token metadata.
5. `imageUrl`: The token's static image url to be used in the metadata.

> [!TIP]
>
> **PRESETS**
>
> On [https://jup.ag/studio](https://jup.ag/studio), you can find a few different presets to get you started.
>
> > [accordion] Meme
> >
> > **Great for memes, similar profile to traditional meme launches.**
> >
> > * People begin buying your token at 16K Market Cap (MC) in USDC.
> > * It graduates to a Meteora pool at 69K MC.
> > * Your pool raises \~17.94K USDC before graduation.
> >
> > ```js theme={null}
> > buildCurveByMarketCapParam: {
> >     quoteMint: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
> >     initialMarketCap: 16000,
> >     migrationMarketCap: 69000,
> >     tokenQuoteDecimal: 6,
> >     lockedVestingParam: {
> >         totalLockedVestingAmount: 0,
> >         cliffUnlockAmount: 0,
> >         numberOfVestingPeriod: 0,
> >         totalVestingDuration: 0,
> >         cliffDurationFromMigrationTime: 0,
> >     },
> > },
> > antiSniping: false,
> > fee: { feeBps: 100, },
> > isLpLocked: true,
> > tokenName: '',
> > tokenSymbol: '',
> > tokenImageContentType: 'image/jpeg',
> > creator: wallet.publicKey.toBase58(),
> > ```
>
> > [accordion] Indie
> >
> > **For projects ready to take it up a notch. More capital required to bond, but you'll have deeper liquidity and more LP fees when you graduate.**
> >
> > * People begin buying your token at 32k Market Cap (MC) in USDC.
> > * It graduates to a Meteora pool at 240k MC.
> > * Your pool raises \~57.78K USDC before graduation.
> > * 10% of total supply will be vested daily over 12 months.
> >
> > ```js theme={null}
> > buildCurveByMarketCapParam: {
> >     quoteMint: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
> >     initialMarketCap: 32000,
> >     migrationMarketCap: 240000,
> >     tokenQuoteDecimal: 6,
> >     lockedVestingParam: {
> >         totalLockedVestingAmount: 100000000,
> >         cliffUnlockAmount: 0,
> >         numberOfVestingPeriod: 365,
> >         totalVestingDuration: 31536000,
> >         cliffDurationFromMigrationTime: 0,
> >     },
> > },
> > antiSniping: true,
> > fee: { feeBps: 100, },
> > isLpLocked: true,
> > tokenName: '',
> > tokenSymbol: '',
> > tokenImageContentType: 'image/jpeg',
> > creator: wallet.publicKey.toBase58(),
> > ```
>
> > [accordion] Custom
> >
> > Just pass in the parameters you need!

```js theme={null}
const createTransaction = await (
    await fetch (
      'https://api.jup.ag/studio/v1/dbc-pool/create-tx',
      {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': 'your-api-key',
        },
        body: JSON.stringify({
            buildCurveByMarketCapParam: {
                quoteMint: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v', // or SOL or JUP
                initialMarketCap: 16000, // This means 16_000 USDC
                migrationMarketCap: 69000, // This means 69_000 USDC
                tokenQuoteDecimal: 6,
                lockedVestingParam: {
                    totalLockedVestingAmount: 0,
                    cliffUnlockAmount: 0,
                    numberOfVestingPeriod: 0,
                    totalVestingDuration: 0,
                    cliffDurationFromMigrationTime: 0,
                },
            },
            antiSniping: true,
            fee: { feeBps: 100, },
            isLpLocked: true,
            tokenName: '',
            tokenSymbol: '',
            tokenImageContentType: 'image/jpeg',
            creator: wallet.publicKey.toBase58(),
        }, null, 2)
    })
).json();
```

## Token Metadata

The following 2 steps, are to upload your token image and metadata to the **static URL**, which will be the URI in the onchain metadata of your token.

Example

* URI/ Off-chain Metadata: `https://static-create.jup.ag/metadata/{mint}.json`
* Image: `https://static-create.jup.ag/images/{mint}`

You can refer to this to understand Token Metadata on Solana: [https://developers.metaplex.com/token-metadata](https://developers.metaplex.com/token-metadata)

### Upload Image

From the response of the `create-tx` endpoint, we will need the `imagePresignedUrl` to make a **`PUT` request** to the url provided, in order to upload the token image.

```js theme={null}
const imageResponse = await fetch(createTransaction.imagePresignedUrl, {
    method: 'PUT',
    headers: {
        'Content-Type': 'image/jpeg', // Adjust based on the image type passed in previously
    },
    body: fs.readFileSync('./token.jpeg'), // Assuming the image file is located in the same folder
});
```

### Upload Metadata

From the response of the `create-tx` endpoint, we will need the `metadataPresignedUrl` to make a **`PUT` request** to the url provided, in order to upload the token metadata.

```js theme={null}
const metadataResponse = await fetch(createTransaction.metadataPresignedUrl, {
    method: 'PUT',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        name: '',
        symbol: '',
        description: '',
        image: createTransaction.imageUrl,
        website: '',
        twitter: '',
        telegram: '',
    }, null, 2),
});
```

## Submit Transaction

After you have uploaded your token image and token metadata, you can proceed to signing and making a post request to the `submit` endpoint - this will allow Jupiter Studio to complete the transaction and submit it to the network on your behalf.

> [!NOTE]
>
> **NOTE**
>
> * Do note that the endpoint expects the `requestBody`'s `content` to be in [`multipart/form-data` format](https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest_API/Using_FormData_Objects).
> * Ensure the file types and size of the image file is manageable.

> [!NOTE]
>
> **NOTE**
>
> The `content` and `headerImage` refers to the Studio dedicated page's token description and header image of the page, they are not on-chain metadata. This is meant for you to customize the Studio dedicated page as you wish - to include lore, story or just a nice looking banner!
>
> The `content` and `headerImage` are stored off-chain for our frontend to ingest and display.
>
> [Do not confuse this with the uploading of token metadata, they are done separately.](#token-metadata)

```js theme={null}
import { VersionedTransaction } from '@solana/web3.js';
import fs from 'fs';

const transaction = VersionedTransaction.deserialize(Buffer.from(createTransaction.transaction, 'base64'));
transaction.sign([wallet]);
const signedTransaction = Buffer.from(transaction.serialize()).toString('base64');

const formData = new FormData();
formData.append('transaction', signedTransaction);
formData.append('owner', wallet.publicKey.toBase58());
formData.append('content', '');
formData.append(
    'headerImage',
    new File(
        [fs.readFileSync('/Path/to/header.jpeg')],
        'header.jpeg',
        { type: 'image/jpeg' },
    )
);

const result = await (
    await fetch (
      'https://api.jup.ag/studio/v1/dbc-pool/submit',
      {
        method: 'POST',
        body: formData,
        headers: {
            'x-api-key': 'your-api-key',
        },
    })
).json();
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.