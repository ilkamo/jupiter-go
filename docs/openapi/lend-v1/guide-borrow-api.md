> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Borrow API (Beta)

> Read borrow vaults and positions, and build collateralised borrow transactions via the Jupiter Lend API.

> [!TIP]
>
> **API REFERENCE**
>
> For the full request and response schemas, see the [Lend API Reference](https://developers.jup.ag/docs/api-reference/lend/borrow/vaults.md).

The Borrow API lets you read borrow vaults and user positions, and build borrow transactions. Borrowing uses a single universal `operate` endpoint: one call can create a position, deposit collateral, borrow, repay, and withdraw, driven by signed `colAmount` and `debtAmount` values. Positions are represented as NFTs.

## Prerequisite

> [accordion] Dependencies
>
> ```bash theme={null}
> npm install @solana/web3.js@1 @solana/spl-token # web3.js v1; spl-token is used to wrap SOL for WSOL-collateral vaults
> ```

> [accordion] RPC
>
> **Set up RPC**
>
> > [!NOTE]
> >
> > **NOTE**
> >
> > Solana provides a [default RPC endpoint](https://solana.com/docs/core/clusters). As your application grows, provision your own or a 3rd party provider's RPC endpoint such as [Helius](https://helius.dev/) or [Triton](https://triton.one/).
>
> ```js theme={null}
> import { Connection } from "@solana/web3.js";
>
> const connection = new Connection("https://api.mainnet-beta.solana.com");
> ```

> [accordion] Wallet
>
> **Set up Development Wallet**
>
> ```js theme={null}
> import { Keypair } from "@solana/web3.js";
> import fs from "fs";
>
> const privateKeyArray = JSON.parse(fs.readFileSync("/Path/To/.config/solana/id.json", "utf8").trim());
> const wallet = Keypair.fromSecretKey(new Uint8Array(privateKeyArray));
> ```

## Markets

All borrow endpoints accept an optional `market` parameter: `main` (default) or `ethena`. Pass it as a query parameter (`?market=ethena`) on any endpoint, including operate calls. Omitting it uses the `main` market.

Vault and position IDs are scoped to a market. The same `vaultId` refers to a different vault in `main` than in `ethena`, so always pass the same `market` to operate that you used to read the vault or position.

## Read vault and position data

Use the read endpoints to discover vaults and inspect user positions. Both accept GET requests.

### Vaults

`GET https://api.jup.ag/lend/v1/borrow/vaults` returns every borrow vault with its collateral token (`supplyToken`), borrow token (`borrowToken`), risk parameters, rates, and available liquidity. The `id` field is the `vaultId` you pass to operate calls.

```js theme={null}
const vaults = await (
    await fetch("https://api.jup.ag/lend/v1/borrow/vaults", {
        headers: { "x-api-key": "your-api-key" },
    })
).json();
```

Each vault reports `collateralFactor` (max LTV, where `800` corresponds to 80%), `liquidationThreshold`, `borrowable` and `withdrawable` (available liquidity in base units), and `minimumBorrowing` (the smallest non-zero debt a position may hold).

### Positions

`GET https://api.jup.ag/lend/v1/borrow/positions?users={user1},{user2}` returns the borrow positions for one or more wallets (comma-separated). Each position reports its NFT `id`, `vaultId`, `supply` (collateral), `borrow` (debt), and `dustBorrow` (residual debt including accrued interest).

```js theme={null}
const positions = await (
    await fetch(
        "https://api.jup.ag/lend/v1/borrow/positions?users={user1},{user2}",
        { headers: { "x-api-key": "your-api-key" } }
    )
).json();
```

## Operate

`POST https://api.jup.ag/lend/v1/borrow/operate` builds a borrow transaction from an `OperatePayload`. The operation performed is determined by the sign of `colAmount` (collateral) and `debtAmount` (debt):

| Action | `colAmount` | `debtAmount` | Notes |
| - | - | - | - |
| Create position + deposit | `> 0` | `"0"` | Set `positionId: 0` |
| Deposit collateral | `> 0` | `"0"` | |
| Borrow | `"0"` | `> 0` | Debt must be `>= minimumBorrowing` |
| Repay | `"0"` | `< 0` | |
| Withdraw collateral | `< 0` | `"0"` | |
| Deposit and borrow | `> 0` | `> 0` | Single transaction |
| Repay all debt | `"0"` | `MIN_I128` | Clears accrued-interest dust |
| Withdraw all collateral | `MIN_I128` | `"0"` | |

`MIN_I128` is the literal string `-170141183460469231731687303715884105728`.

The `OperatePayload` fields:

| Field | Type | Required | Description |
| - | - | - | - |
| `vaultId` | `number` | Yes | Vault ID from `GET /borrow/vaults`. |
| `positionId` | `number` | Yes | Position NFT ID to operate on. Use `0` to create a new position. |
| `positionOwner` | `string` | No | Owner of the position. Defaults to `signer`. |
| `signer` | `string` | Yes | Wallet that signs and pays for the transaction. |
| `colAmount` | `string` | Yes | Signed collateral amount in supply-token base units. |
| `debtAmount` | `string` | Yes | Signed debt amount in borrow-token base units. |

```js theme={null}
const operateResponse = await (
    await fetch("https://api.jup.ag/lend/v1/borrow/operate", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "x-api-key": "your-api-key",
        },
        body: JSON.stringify({
            vaultId: 1,
            positionId: 0, // 0 creates a new position
            signer: wallet.publicKey.toBase58(),
            colAmount: "30000000", // deposit 0.03 SOL of collateral
            debtAmount: "0",
        }),
    })
).json();

// { nftId: 9062, transaction: "<base64 unsigned VersionedTransaction>" }
```

The response includes `nftId` (the position NFT ID, newly assigned when `positionId` was `0`) and a base64-encoded unsigned `VersionedTransaction`.

The API creates the borrow-token associated token account for the `signer` when it does not already exist, so borrowing to a fresh wallet works without adding a create-account instruction. You do not need to pre-create the output token account.

### Native SOL (WSOL) handling

The API does not wrap or unwrap native SOL. This matters whenever a vault's `supplyToken` or `borrowToken` is `So11111111111111111111111111111111111111112` (WSOL):

* **Depositing WSOL collateral:** wrap SOL into your WSOL token account first. Without a funded WSOL account, the deposit fails with a token-program `insufficient funds` error. Wrap slightly more than you deposit: depositing your exact WSOL balance can fail because the transfer rounds up by a few base units.
* **Borrowing or withdrawing WSOL:** you receive wrapped SOL in your WSOL token account. Close the account yourself if you want native SOL back.

> [!NOTE]
>
> Repaying the exact `borrow` amount can leave interest dust below `minimumBorrowing`, which fails with `VaultUserDebtTooLow`. Repay with `debtAmount: MIN_I128` to clear all debt before closing a position.

## Build your own transaction

The Borrow API provides two ways to build an operate transaction. Use `/operate` for a ready-to-sign transaction, or `/operate-instructions` when you need the raw instructions to compose with other instructions or for CPI.

### Transaction

Call `/operate` (above), deserialize the base64 transaction, sign, and send.

```js theme={null}
import { VersionedTransaction } from "@solana/web3.js";

const transaction = VersionedTransaction.deserialize(
    Buffer.from(operateResponse.transaction, "base64")
);
transaction.sign([wallet]);
const signature = await connection.sendRawTransaction(transaction.serialize(), {
    skipPreflight: false,
    maxRetries: 5,
    preflightCommitment: "confirmed",
});
console.log(`https://solscan.io/tx/${signature}`);
```

### Instructions

`POST https://api.jup.ag/lend/v1/borrow/operate-instructions` takes the same `OperatePayload` and returns `{ nftId, instructions, addressLookupTableAddresses }`. Resolve the address lookup tables and compile a v0 transaction.

```js theme={null}
import {
    PublicKey,
    TransactionMessage,
    TransactionInstruction,
    VersionedTransaction,
} from "@solana/web3.js";

const borrowIx = await (
    await fetch("https://api.jup.ag/lend/v1/borrow/operate-instructions", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "x-api-key": "your-api-key",
        },
        body: JSON.stringify({
            vaultId: 1,
            positionId: 9062, // existing position NFT id
            positionOwner: wallet.publicKey.toBase58(),
            signer: wallet.publicKey.toBase58(),
            colAmount: "0",
            debtAmount: "1100000", // borrow 1.1 USDC
        }),
    })
).json();

const deserializeInstruction = (instruction) =>
    new TransactionInstruction({
        programId: new PublicKey(instruction.programId),
        keys: instruction.accounts.map((key) => ({
            pubkey: new PublicKey(key.pubkey),
            isSigner: key.isSigner,
            isWritable: key.isWritable,
        })),
        data: Buffer.from(instruction.data, "base64"),
    });

// Resolve the address lookup tables returned by the API
const lookupTables = (
    await Promise.all(
        (borrowIx.addressLookupTableAddresses ?? []).map((addr) =>
            connection.getAddressLookupTable(new PublicKey(addr))
        )
    )
).map((res) => res.value).filter(Boolean);

const blockhash = (await connection.getLatestBlockhash()).blockhash;
const messageV0 = new TransactionMessage({
    payerKey: wallet.publicKey,
    recentBlockhash: blockhash,
    instructions: borrowIx.instructions.map(deserializeInstruction),
}).compileToV0Message(lookupTables);

const transaction = new VersionedTransaction(messageV0);
transaction.sign([wallet]);
const signature = await connection.sendRawTransaction(transaction.serialize(), {
    skipPreflight: false,
    maxRetries: 5,
    preflightCommitment: "confirmed",
});
console.log(`https://solscan.io/tx/${signature}`);
```

### CPI

* Refer to [github.com/jup-ag/jupiter-lend/blob/main/docs/borrow/cpi.md](https://github.com/jup-ag/jupiter-lend/blob/main/docs/borrow/cpi.md) for the CPI example
* Refer to [github.com/jup-ag/jupiter-lend/blob/main/target/idl/vaults.json](https://github.com/jup-ag/jupiter-lend/blob/main/target/idl/vaults.json) for the IDL

## Full code example

Create a position, deposit collateral, then borrow against it. The borrow step uses `/operate-instructions` to show the address lookup table flow.

```typescript expandable theme={null}
import {
    Connection,
    Keypair,
    PublicKey,
    SystemProgram,
    TransactionMessage,
    TransactionInstruction,
    VersionedTransaction,
} from "@solana/web3.js";
import {
    NATIVE_MINT,
    getAssociatedTokenAddressSync,
    createAssociatedTokenAccountIdempotentInstruction,
    createSyncNativeInstruction,
} from "@solana/spl-token";
import fs from "fs";

const API = "https://api.jup.ag/lend/v1";
const API_KEY = "your-api-key";
const VAULT_ID = 1; // WSOL collateral -> USDC borrow
const COLLATERAL_LAMPORTS = 30_000_000; // 0.03 SOL of WSOL collateral

// 1. Load wallet and connection
const wallet = Keypair.fromSecretKey(
    new Uint8Array(JSON.parse(fs.readFileSync("/path/to/keypair.json", "utf8").trim()))
);
const connection = new Connection("https://api.mainnet-beta.solana.com", "confirmed");
const signer = wallet.publicKey.toBase58();

const headers = { "Content-Type": "application/json", "x-api-key": API_KEY };

async function signSend(b64: string, label: string) {
    const tx = VersionedTransaction.deserialize(Buffer.from(b64, "base64"));
    tx.sign([wallet]);
    const sig = await connection.sendRawTransaction(tx.serialize(), {
        skipPreflight: false,
        maxRetries: 5,
        preflightCommitment: "confirmed",
    });
    await connection.confirmTransaction(sig, "confirmed");
    console.log(`${label}: https://solscan.io/tx/${sig}`);
    return sig;
}

// 2. Wrap SOL into WSOL. Vault 1 uses WSOL collateral and the API does not
//    wrap native SOL for you. Wrap slightly more than you deposit to cover rounding.
const wsolAta = getAssociatedTokenAddressSync(NATIVE_MINT, wallet.publicKey);
const wrapTx = new VersionedTransaction(
    new TransactionMessage({
        payerKey: wallet.publicKey,
        recentBlockhash: (await connection.getLatestBlockhash()).blockhash,
        instructions: [
            createAssociatedTokenAccountIdempotentInstruction(wallet.publicKey, wsolAta, wallet.publicKey, NATIVE_MINT),
            SystemProgram.transfer({ fromPubkey: wallet.publicKey, toPubkey: wsolAta, lamports: COLLATERAL_LAMPORTS + 10_000 }),
            createSyncNativeInstruction(wsolAta),
        ],
    }).compileToV0Message()
);
wrapTx.sign([wallet]);
const wrapSig = await connection.sendRawTransaction(wrapTx.serialize(), {
    skipPreflight: false,
    maxRetries: 5,
    preflightCommitment: "confirmed",
});
await connection.confirmTransaction(wrapSig, "confirmed");
console.log(`wrap: https://solscan.io/tx/${wrapSig}`);

// 3. Create position + deposit collateral via /operate (positionId: 0)
const deposit = await (
    await fetch(`${API}/borrow/operate`, {
        method: "POST",
        headers,
        body: JSON.stringify({
            vaultId: VAULT_ID,
            positionId: 0,
            signer,
            colAmount: String(COLLATERAL_LAMPORTS), // 0.03 WSOL (wrapped above)
            debtAmount: "0",
        }),
    })
).json();
const nftId = deposit.nftId;
await signSend(deposit.transaction, "deposit");

// 4. Borrow against the position via /operate-instructions
const borrow = await (
    await fetch(`${API}/borrow/operate-instructions`, {
        method: "POST",
        headers,
        body: JSON.stringify({
            vaultId: VAULT_ID,
            positionId: nftId,
            positionOwner: signer,
            signer,
            colAmount: "0",
            debtAmount: "1100000", // 1.1 USDC
        }),
    })
).json();

const deserialize = (ix: any) =>
    new TransactionInstruction({
        programId: new PublicKey(ix.programId),
        keys: ix.accounts.map((k: any) => ({
            pubkey: new PublicKey(k.pubkey),
            isSigner: k.isSigner,
            isWritable: k.isWritable,
        })),
        data: Buffer.from(ix.data, "base64"),
    });

const lookupTables = (
    await Promise.all(
        (borrow.addressLookupTableAddresses ?? []).map((a: string) =>
            connection.getAddressLookupTable(new PublicKey(a))
        )
    )
).map((r) => r.value).filter(Boolean) as any[];

const blockhash = (await connection.getLatestBlockhash()).blockhash;
const message = new TransactionMessage({
    payerKey: wallet.publicKey,
    recentBlockhash: blockhash,
    instructions: borrow.instructions.map(deserialize),
}).compileToV0Message(lookupTables);
const borrowTx = new VersionedTransaction(message);
borrowTx.sign([wallet]);
const sig = await connection.sendRawTransaction(borrowTx.serialize(), {
    skipPreflight: false,
    maxRetries: 5,
    preflightCommitment: "confirmed",
});
await connection.confirmTransaction(sig, "confirmed");
console.log(`borrow: https://solscan.io/tx/${sig}`);
```

## Related

> [card] [Borrow API Reference](https://developers.jup.ag/docs/api-reference/lend/borrow/vaults.md)
>
> Full request and response schemas for vaults, positions, operate, and operate-instructions.

> [card] [Read SDK](https://developers.jup.ag/docs/lend/borrow/read-vault-data.md)
>
> Read vault config, state, and user positions with `@jup-ag/lend-read`.

> [card] [Borrow SDK operations](https://developers.jup.ag/docs/lend/borrow/create-position.md)
>
> Build operate flows (create, deposit, borrow, repay, withdraw) with `@jup-ag/lend`.

> [card] [Borrow CPI](https://developers.jup.ag/docs/lend/borrow/cpi.md)
>
> Cross-program invocation for vault operations.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.