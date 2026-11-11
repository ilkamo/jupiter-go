> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Send (Beta)

> Build a Send transaction that transfers tokens to a recipient via an invite code.

This guide walks through crafting a Send transaction using the `/send/v1/craft-send` endpoint. You will generate an invite code, derive a recipient keypair, POST to the endpoint, then sign with both sender and recipient keypairs before submitting.

## Overview

**Step 1**

Create invite code.

**Step 2**

From utils, derive the secret key - a deterministic 64-byte Solana secret key (32 bytes private + 32 bytes public key).

**Step 3**

Create Solana Keypair instance from the secret key.

**Step 4**

Post request to get Send transaction.

**Step 5**

Sign with both sender and recipient keypair, then send transaction and wait for confirmation.

> [!NOTE]
>
> **NOTE**
>
> [Please ensure that you have set up the prerequisites](https://developers.jup.ag/docs/send/invite-code.md#overview).

> [accordion] Full Code Snippet
>
> ```js theme={null}
> import { create_invite_code, invite_code_to_priv_key } from "./utils.js";
> import {
>   Connection,
>   Keypair,
>   VersionedTransaction,
> } from "@solana/web3.js";
> import fs from "fs";
>
> const connection = new Connection('insert-rpc');
> const senderPrivateKey = JSON.parse(fs.readFileSync('/Path/to/sender/id.json', 'utf8').trim());
> const sender = Keypair.fromSecretKey(new Uint8Array(senderPrivateKey));
>
> // STEP 1: Create 12-character invite code
> const invite_code = await create_invite_code();
>
> // STEP 2: Derive secret key (public and private key)
> const secret_key = invite_code_to_priv_key(invite_code);
>
> // STEP 3: Use secret key to create Solana Keypair instance
> const recipient = Keypair.fromSecretKey(secret_key);
>
> // STEP 4: Post request for a Send transaction
> const craftSendTransaction = await (
>     await fetch ('https://api.jup.ag/send/v1/craft-send', {
>         method: 'POST',
>         headers: {
>             'Content-Type': 'application/json',
>             'x-api-key': 'your-api-key',
>         },
>         body: JSON.stringify({
>             inviteSigner: recipient.publicKey.toBase58(),
>             sender: sender.publicKey.toBase58(),
>             amount: "10000000", // atomic amount before decimals
>             // mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // Defaults to SOL if `mint` is not provided
>         }, null, 2)
>     })
> ).json();
>
> // STEP 5: Use sender and receipient keypair to sign and send to network
> const transaction = VersionedTransaction.deserialize(Buffer.from(craftSendTransaction.tx, 'base64'));
> transaction.sign([sender, recipient]); // SIGN with both SENDER and RECIPIENT keypair
> const transactionBinary = transaction.serialize();
> const blockhashInfo = await connection.getLatestBlockhashAndContext({ commitment: "confirmed" });
>
> const signature = await connection.sendRawTransaction(transactionBinary, {
>   maxRetries: 0,
>   skipPreflight: true,
> });
>
> // Log the signature immediately after sending, before confirmation
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
>   };
> } catch (error) {
>   console.error(`Error confirming transaction: ${error}`);
>   console.log(`Examine the transaction status: https://solscan.io/tx/${signature}`);
> };
> ```

## Imports

```js theme={null}
import { create_invite_code, invite_code_to_priv_key } from "./utils.js";
import {
  Connection,
  Keypair,
} from "@solana/web3.js";
import fs from "fs";

const connection = new Connection('insert-rpc');
const senderPrivateKey = JSON.parse(fs.readFileSync('/Path/to/sender/id.json', 'utf8').trim());
const sender = Keypair.fromSecretKey(new Uint8Array(senderPrivateKey));
```

## Create Invite Code

```js theme={null}
// STEP 1: Create 12-character invite code
const invite_code = await create_invite_code();

// STEP 2: Derive secret key (public and private key)
const secret_key = invite_code_to_priv_key(invite_code);

// STEP 3: Use secret key to create Solana Keypair instance
const recipient = Keypair.fromSecretKey(secret_key);
```

## Craft Send

> [!NOTE]
>
> **API PARAMS**
>
> * The `amount` is in its atomic value before applying decimals, e.g. 1 USDC is 1\_000\_000.
> * The `mint` defaults to SOL if not provided, if provided it can be any token mint.

> [!NOTE]
>
> **SIGNING AND SENDING**
>
> * After getting the transaction, you need to sign with **both sender and recipient** keypair.
> * You can send the transaction to the network via any method.

```js theme={null}
// STEP 4: Post request for a Send transaction
const craftSendTransaction = await (
    await fetch ('https://api.jup.ag/send/v1/craft-send', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': 'your-api-key',
        },
        body: JSON.stringify({
            inviteSigner: recipient.publicKey.toBase58(),
            sender: sender.publicKey.toBase58(),
            amount: "10000000",
            // mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        }, null, 2)
    })
).json();
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.