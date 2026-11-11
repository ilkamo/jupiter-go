> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Clawback (Beta)

> Reclaim tokens from an unclaimed Send invite before it expires.

This guide walks through crafting a Clawback transaction using the `/send/v1/craft-clawback` endpoint. You will load an existing invite code, derive the Program Derived Address (PDA), POST to the endpoint, then sign with the sender keypair to reclaim unclaimed tokens.

## Overview

**Step 1**

Load invite code.

**Step 2**

Load public key from invite.

**Step 3**

Find the [Program Derived Address (PDA)](https://solana.com/core/pda) of the invite.

* Uses `"invite"` and the public key of recipient at seed.

**Step 4**

Post request to get Clawback transaction.

**Step 5**

Sign with sender keypair, then send transaction and wait for confirmation.

> [!NOTE]
>
> **NOTE**
>
> [Please ensure that you have set up the prerequisites](https://developers.jup.ag/docs/send/invite-code.md#overview).

> [accordion] Full Code Snippet
>
> ```js theme={null}
> import { invite_code_to_priv_key } from "./utils.js";
> import {
>   Connection,
>   Keypair,
>   PublicKey,
>   VersionedTransaction,
> } from "@solana/web3.js";
> import fs from "fs";
>
> const connection = new Connection('insert-rpc');
> const senderPrivateKey = JSON.parse(fs.readFileSync('/Path/to/sender/id.json', 'utf8').trim());
> const sender = Keypair.fromSecretKey(new Uint8Array(senderPrivateKey));
> process.loadEnvFile('.env');
>
> // STEP 1: Load invite code
> const invite_code = process.env.INVITE_CODE;
>
> // STEP 2: Load the public key from the invite code
> const secret_key = invite_code_to_priv_key(invite_code);
> const pubkey = Keypair.fromSecretKey(secret_key).publicKey;
>
> // STEP 3: Find the Program Derived Address (PDA) for the invite
> // Uses `"invite"` as seed + the public key
> // PDAs are deterministic addresses owned by the program
> const invite_pda = PublicKey.findProgramAddressSync(
>     [Buffer.from("invite"), pubkey.toBuffer()],
>     new PublicKey("inv1tEtSwRMtM44tbvJGNiTxMvDfPVnX9StyqXfDfks")
>   )[0];
>
> // STEP 4: Post request for a Clawback transaction
> const craftClawbackTransaction = await (
>     await fetch ('https://api.jup.ag/send/v1/craft-clawback', {
>         method: 'POST',
>         headers: {
>             'Content-Type': 'application/json',
>             'x-api-key': 'your-api-key',
>         },
>         body: JSON.stringify({
>             invitePDA: invite_pda.toBase58(),
>             sender: sender.publicKey.toBase58(),
>         }, null, 2)
>     })
> ).json();
>
> // STEP 5: Use sender keypair to sign and send to network
> const transaction = VersionedTransaction.deserialize(Buffer.from(craftClawbackTransaction.tx, 'base64'));
> transaction.sign([sender]); // SIGN with SENDER
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
>   }
> } catch (error) {
>   console.error(`Error confirming transaction: ${error}`);
>   console.log(`Examine the transaction status: https://solscan.io/tx/${signature}`);
> }
> ```

## Imports

```js theme={null}
import { invite_code_to_priv_key } from "./utils.js";
import {
  Connection,
  Keypair,
  PublicKey,
  VersionedTransaction,
} from "@solana/web3.js";
import fs from "fs";

const connection = new Connection('insert-rpc');
const senderPrivateKey = JSON.parse(fs.readFileSync('/Path/to/sender/id.json', 'utf8').trim());
const sender = Keypair.fromSecretKey(new Uint8Array(senderPrivateKey));
process.loadEnvFile('.env');
```

## Invite Code and Public Key

```js theme={null}
// STEP 1: Load invite code
const invite_code = process.env.INVITE_CODE;

// STEP 2: Load the public key from the invite code
const secret_key = invite_code_to_priv_key(invite_code); // Follow the utils.js guide
const pubkey = Keypair.fromSecretKey(secret_key).publicKey;
```

## Invite PDA

```js theme={null}
// STEP 3: Find the Program Derived Address (PDA) for the invite
// Uses `"invite"` as seed + the public key
// PDAs are deterministic addresses owned by the program
const invite_pda = PublicKey.findProgramAddressSync(
    [Buffer.from("invite"), pubkey.toBuffer()],
    new PublicKey("inv1tEtSwRMtM44tbvJGNiTxMvDfPVnX9StyqXfDfks")
  )[0];
```

## Craft Clawback

> [!NOTE]
>
> **NOTE**
>
> The clawback will return the full amount including leftover transaction fees and/or rent back to the sender.

```js theme={null}
// STEP 4: Post request for a Clawback transaction
const craftClawbackTransaction = await (
    await fetch ('https://api.jup.ag/send/v1/craft-clawback', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': 'your-api-key',
        },
        body: JSON.stringify({
            invitePDA: invite_pda.toBase58(),
            sender: sender.publicKey.toBase58(),
        }, null, 2)
    })
).json();
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.