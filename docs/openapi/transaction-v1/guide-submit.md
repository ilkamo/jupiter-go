> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Transaction Submission

> Submit signed transactions to the Solana cluster through Jupiter's landing infrastructure by pointing a Solana RPC client at tx.jup.ag

`tx.jup.ag` is a Solana RPC-compatible endpoint that forwards your signed transactions to the Solana cluster through Jupiter's proprietary landing infrastructure, the same stack that powers Jupiter's own swap products. Point any Solana client at it, authenticate with your Jupiter API key, and call `sendTransaction` as you would against any RPC node. It accepts any valid signed Solana transaction with a SOL tip.

## How it works

`tx.jup.ag` is Solana RPC-compatible where possible: it implements the standard [`sendTransaction`](https://solana.com/docs/rpc/http/sendtransaction) method, so any Solana client library can talk to it with only an endpoint change. It is send-only, so it does not serve `getLatestBlockhash`, `simulateTransaction`, or confirmation queries; keep your usual RPC for those.

1. Build your transaction (using `/build`, or assemble your own)
2. Include a SOL tip transfer instruction (minimum 0.001 SOL) to one of 16 tip receiver accounts
3. Sign the transaction
4. Send it with `sendTransaction` to `https://tx.jup.ag`, passing your API key in the `x-api-key` header

For `/build` transactions, use the `tipAmount` parameter to have the tip instruction included automatically. For non-Jupiter transactions, add a standard SOL transfer instruction to one of the [tip receiver accounts](#adding-tip-instruction-manually).

The request is a standard JSON-RPC call, so any language works:

```bash theme={null}
curl https://tx.jup.ag \
  -H "Content-Type: application/json" \
  -H "x-api-key: $JUPITER_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "sendTransaction",
    "params": ["<base64-signed-transaction>", { "encoding": "base64" }]
  }'
```

A successful response returns the signature in `result`. This means the transaction was accepted and forwarded, not that it landed, so confirm on your own RPC. A transaction without a valid tip is rejected with `Transaction must include a Jupiter tip instruction`.

## Requirements

| Requirement | Details |
| - | - |
| **Endpoint** | `POST https://tx.jup.ag`, JSON-RPC method `sendTransaction` |
| **Valid signed transaction** | Must be a valid Solana transaction with all required signatures |
| **SOL tip** | Minimum 1,000,000 lamports (0.001 SOL) transferred to one of the 16 tip receiver accounts |
| **Transaction size** | Must not exceed the Solana transaction size limit (1232 bytes) |
| **API access** | Requires a Jupiter API key in the `x-api-key` header |

The `sendTransaction` config (the second param) takes:

| Field | Value |
| - | - |
| `encoding` | Defaults to `base64` (the only supported value) |
| `skipPreflight` | Preflight is not supported; defaults to `true`. An explicit `false` is rejected with `-1015` |
| `maxRetries` | Defaults to `0`. A non-zero value is rejected with `-1015` |
| `swqosOnly` | Optional. `false` (default) uses SWQoS + Jito; `true` uses SWQoS only. See [Tips and fees](#tips-and-fees) |

## Tips and fees

Every transaction must include the **Jupiter tip** instruction (minimum 0.001 SOL) or it is rejected. The Jupiter tip is flat: tipping more does not improve landing. Add **priority fees** when the network is congested.

Whether to add a **Jito tip** depends on `swqosOnly`:

| `swqosOnly` | Routing | Jito tip |
| - | - | - |
| `false` (default) | SWQoS + Jito | Not needed, Beam adds a small one. Optionally add your own to boost Jito landing. |
| `true` | SWQoS only | Do not add, it will not help. |

## Code example

### Using /build

Use [`/build`](https://developers.jup.ag/docs/swap/build) with the `tipAmount` parameter to automatically include a tip instruction in the response.

```typescript @solana/web3.js expandable theme={null}
import {
  AddressLookupTableAccount,
  ComputeBudgetProgram,
  Connection,
  Keypair,
  PublicKey,
  TransactionInstruction,
  TransactionMessage,
  VersionedTransaction,
} from "@solana/web3.js";
import bs58 from "bs58";

// ── Types matching the /build response ───────────────────────────────────────

type ApiAccount = { pubkey: string; isSigner: boolean; isWritable: boolean };

type ApiInstruction = {
  programId: string;
  accounts: ApiAccount[];
  data: string; // base64
};

type BuildResponse = {
  computeBudgetInstructions: ApiInstruction[];
  setupInstructions: ApiInstruction[];
  swapInstruction: ApiInstruction;
  cleanupInstruction: ApiInstruction | null;
  otherInstructions: ApiInstruction[];
  tipInstruction: ApiInstruction;
  addressesByLookupTableAddress: Record<string, string[]> | null;
  blockhashWithMetadata: {
    blockhash: number[];
    lastValidBlockHeight: number;
  };
};

// ── Helpers ──────────────────────────────────────────────────────────────────

const CU_LIMIT_MAX = 1_400_000;

function toInstruction(ix: ApiInstruction): TransactionInstruction {
  return new TransactionInstruction({
    programId: new PublicKey(ix.programId),
    keys: ix.accounts.map((acc) => ({
      pubkey: new PublicKey(acc.pubkey),
      isSigner: acc.isSigner,
      isWritable: acc.isWritable,
    })),
    data: Buffer.from(ix.data, "base64"),
  });
}

// ── Main ─────────────────────────────────────────────────────────────────────

const API_KEY = process.env.JUPITER_API_KEY!;
const connection = new Connection(process.env.RPC_URL!);
// Send-only endpoint; blockhash and confirmation stay on your own RPC.
const txConnection = new Connection("https://tx.jup.ag", {
  httpHeaders: { "x-api-key": API_KEY },
});
const signer = Keypair.fromSecretKey(
  bs58.decode(process.env.BS58_PRIVATE_KEY!),
);

// 1. Call /build with your swap parameters
const buildRes = await fetch(
  "https://api.jup.ag/swap/v2/build?" +
    new URLSearchParams({
      inputMint: "So11111111111111111111111111111111111111112",
      outputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
      amount: "100000000",
      taker: signer.publicKey.toString(),
      tipAmount: "1000000",
    }),
  { headers: { "x-api-key": API_KEY } },
);
const build: BuildResponse = await buildRes.json();

// 2. Collect instructions (excluding compute budget — we handle CU limit ourselves)
const instructions = [
  ...build.setupInstructions.map(toInstruction),
  toInstruction(build.swapInstruction),
  ...(build.cleanupInstruction
    ? [toInstruction(build.cleanupInstruction)]
    : []),
  ...build.otherInstructions.map(toInstruction),
  toInstruction(build.tipInstruction),
];

// 3. Prepare blockhash and address lookup tables
const addressLookupTableAccounts: AddressLookupTableAccount[] = [];
for (const addr of Object.keys(build.addressesByLookupTableAddress ?? {})) {
  const result = await connection.getAddressLookupTable(new PublicKey(addr));
  if (result.value) addressLookupTableAccounts.push(result.value);
}

const recentBlockhash = bs58.encode(
  Buffer.from(build.blockhashWithMetadata.blockhash),
);
const { lastValidBlockHeight } = build.blockhashWithMetadata;

// 4. Simulate to estimate CU usage, then apply 1.2x buffer
const simMessage = new TransactionMessage({
  payerKey: signer.publicKey,
  recentBlockhash,
  instructions: [
    ComputeBudgetProgram.setComputeUnitLimit({ units: CU_LIMIT_MAX }),
    ...instructions,
  ],
}).compileToV0Message(addressLookupTableAccounts);

const sim = await connection.simulateTransaction(
  new VersionedTransaction(simMessage),
  { sigVerify: false, replaceRecentBlockhash: true },
);

if (sim.value.err) {
  console.error("Simulation failed:", sim.value.err);
  process.exit(1);
}

const estimatedCUL = sim.value.unitsConsumed
  ? Math.min(Math.ceil(sim.value.unitsConsumed * 1.2), CU_LIMIT_MAX)
  : CU_LIMIT_MAX;

// 5. Build final transaction with estimated CU limit + CU price from response
const messageV0 = new TransactionMessage({
  payerKey: signer.publicKey,
  recentBlockhash,
  instructions: [
    ComputeBudgetProgram.setComputeUnitLimit({ units: estimatedCUL }),
    ...build.computeBudgetInstructions.map(toInstruction),
    ...instructions,
  ],
}).compileToV0Message(addressLookupTableAccounts);

const transaction = new VersionedTransaction(messageV0);
transaction.sign([signer]);

// 6. Submit the signed transaction through tx.jup.ag (or use your own RPC / transaction pipeline)
const signature = await txConnection.sendRawTransaction(
  transaction.serialize(),
  { skipPreflight: true, maxRetries: 0 },
);
console.log("Submitted:", `https://solscan.io/tx/${signature}`);

// 7. Confirm the transaction landed
const confirmation = await connection.confirmTransaction(
  { signature, blockhash: recentBlockhash, lastValidBlockHeight },
  "confirmed",
);

if (confirmation.value.err) {
  console.error("Transaction failed:", confirmation.value.err);
  process.exit(1);
}

console.log("Confirmed:", signature);
```

```typescript @solana/kit expandable theme={null}
import {
  AccountRole,
  Address,
  address,
  AddressesByLookupTableAddress,
  appendTransactionMessageInstructions,
  Base64EncodedBytes,
  Blockhash,
  compileTransaction,
  compressTransactionMessageUsingAddressLookupTables,
  createDefaultRpcTransport,
  createKeyPairSignerFromBytes,
  createSolanaRpc,
  createSolanaRpcFromTransport,
  createTransactionMessage,
  getBase58Decoder,
  getBase58Encoder,
  getBase64Codec,
  getBase64EncodedWireTransaction,
  AccountMeta,
  Instruction,
  pipe,
  setTransactionMessageFeePayer,
  setTransactionMessageLifetimeUsingBlockhash,
  signTransaction,
} from "@solana/kit";

// ── Types matching the /build response ───────────────────────────────────────

type Account = { pubkey: Address; isSigner: boolean; isWritable: boolean };

type ApiInstruction = {
  programId: Address;
  accounts: Account[];
  data: Base64EncodedBytes;
};

type BuildResponse = {
  computeBudgetInstructions: ApiInstruction[];
  setupInstructions: ApiInstruction[];
  swapInstruction: ApiInstruction;
  cleanupInstruction: ApiInstruction | null;
  otherInstructions: ApiInstruction[];
  tipInstruction: ApiInstruction;
  addressesByLookupTableAddress: Record<string, string[]> | null;
  blockhashWithMetadata: {
    blockhash: number[];
    lastValidBlockHeight: number;
  };
};

// ── Helpers ──────────────────────────────────────────────────────────────────

const COMPUTE_BUDGET_PROGRAM = address(
  "ComputeBudget111111111111111111111111111111",
);
const CU_LIMIT_MAX = 1_400_000;

function createInstruction(ix: ApiInstruction): Instruction {
  return {
    programAddress: ix.programId,
    accounts: ix.accounts.map((acc) => ({
      address: acc.pubkey,
      role: acc.isSigner && acc.isWritable
        ? AccountRole.WRITABLE_SIGNER
        : acc.isSigner
          ? AccountRole.READONLY_SIGNER
          : acc.isWritable
            ? AccountRole.WRITABLE
            : AccountRole.READONLY,
    })),
    data: Uint8Array.from(getBase64Codec().encode(ix.data)),
  };
}

function makeSetComputeUnitLimitIx(units: number): ApiInstruction {
  const data = Buffer.alloc(5);
  data.writeUInt8(0x02, 0);
  data.writeUInt32LE(units, 1);
  return {
    programId: COMPUTE_BUDGET_PROGRAM,
    accounts: [],
    data: data.toString("base64") as ApiInstruction["data"],
  };
}

function transformBlockhash(meta: BuildResponse["blockhashWithMetadata"]) {
  return {
    blockhash: getBase58Decoder().decode(
      Uint8Array.from(meta.blockhash),
    ) as Blockhash,
    lastValidBlockHeight: BigInt(meta.lastValidBlockHeight),
  };
}

function transformALTs(
  raw: Record<string, string[]> | null,
): AddressesByLookupTableAddress {
  if (!raw) return {};
  return Object.fromEntries(
    Object.entries(raw).map(([key, addrs]) => [
      address(key),
      addrs.map((a) => address(a)),
    ]),
  );
}

function buildTransaction(
  ixs: Instruction[],
  blockhash: { blockhash: Blockhash; lastValidBlockHeight: bigint },
  alts: AddressesByLookupTableAddress,
  feePayer: Address,
) {
  return pipe(
    createTransactionMessage({ version: 0 }),
    (msg) => appendTransactionMessageInstructions(ixs, msg),
    (msg) => compressTransactionMessageUsingAddressLookupTables(msg, alts),
    (msg) => setTransactionMessageFeePayer(feePayer, msg),
    (msg) => setTransactionMessageLifetimeUsingBlockhash(blockhash, msg),
    (msg) => compileTransaction(msg),
  );
}

// ── Main ─────────────────────────────────────────────────────────────────────

const API_KEY = process.env.JUPITER_API_KEY!;
const rpc = createSolanaRpc(process.env.RPC_URL!);
// Send-only endpoint; blockhash and confirmation stay on your own RPC.
const txRpc = createSolanaRpcFromTransport(
  createDefaultRpcTransport({
    url: "https://tx.jup.ag",
    headers: { "x-api-key": API_KEY },
  }),
);
const signer = await createKeyPairSignerFromBytes(
  getBase58Encoder().encode(process.env.BS58_PRIVATE_KEY!),
);

// 1. Call /build with your swap parameters
const buildRes = await fetch(
  "https://api.jup.ag/swap/v2/build?" +
    new URLSearchParams({
      inputMint: "So11111111111111111111111111111111111111112",
      outputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
      amount: "100000000",
      taker: signer.address,
      tipAmount: "1000000",
    }),
  { headers: { "x-api-key": API_KEY } },
);
const build: BuildResponse = await buildRes.json();

// 2. Collect instructions (excluding compute budget — we handle CU limit ourselves)
const instructions = [
  ...build.setupInstructions.map(createInstruction),
  createInstruction(build.swapInstruction),
  ...(build.cleanupInstruction
    ? [createInstruction(build.cleanupInstruction)]
    : []),
  ...build.otherInstructions.map(createInstruction),
  createInstruction(build.tipInstruction),
];

// 3. Prepare blockhash and address lookup tables
const blockhash = transformBlockhash(build.blockhashWithMetadata);
const alts = transformALTs(build.addressesByLookupTableAddress);

// 4. Simulate to estimate CU usage, then apply 1.2x buffer
const simTx = buildTransaction(
  [createInstruction(makeSetComputeUnitLimitIx(CU_LIMIT_MAX)), ...instructions],
  blockhash,
  alts,
  signer.address,
);
const sim = await rpc
  .simulateTransaction(getBase64EncodedWireTransaction(simTx), {
    encoding: "base64",
    commitment: "confirmed",
    replaceRecentBlockhash: true,
  })
  .send();

if (sim.value.err) {
  console.error("Simulation failed:", sim.value.err);
  process.exit(1);
}

const estimatedCUL = sim.value.unitsConsumed
  ? Math.min(
      Math.ceil(Number(sim.value.unitsConsumed) * 1.2),
      CU_LIMIT_MAX,
    )
  : CU_LIMIT_MAX;

// 5. Build final transaction with estimated CU limit + CU price from response
const compiledTx = buildTransaction(
  [
    createInstruction(makeSetComputeUnitLimitIx(estimatedCUL)),
    ...build.computeBudgetInstructions.map(createInstruction),
    ...instructions,
  ],
  blockhash,
  alts,
  signer.address,
);

// 6. Sign and submit through tx.jup.ag (or use your own RPC / transaction pipeline)
const signedTx = await signTransaction([signer.keyPair], compiledTx);

const signature = await txRpc
  .sendTransaction(getBase64EncodedWireTransaction(signedTx), {
    encoding: "base64",
    skipPreflight: true,
    maxRetries: 0,
  })
  .send();
console.log("Submitted:", `https://solscan.io/tx/${signature}`);

// 7. Confirm the transaction landed
const confirmation = await rpc
  .confirmTransaction(signature, {
    strategy: { type: "blockhash", ...blockhash },
    commitment: "confirmed",
  })
  .send();

if (confirmation.value.err) {
  console.error("Transaction failed:", confirmation.value.err);
  process.exit(1);
}

console.log("Confirmed:", signature);
```

### Adding tip instruction manually

For non-Jupiter transactions (another swap protocol, a program call, a transfer), add a standard SOL transfer to one of the 16 tip receiver accounts. Randomise which account you send to on each transaction to reduce write-lock contention.

```typescript @solana/web3.js expandable theme={null}
import { SystemProgram, PublicKey } from "@solana/web3.js";

const TIP_ACCOUNTS = [
  "GGztQqQ6pCPaJQnNpXBgELr5cs3WwDakRbh1iEMzjgSJ",
  "2MFoS3MPtvyQ4Wh4M9pdfPjz6UhVoNbFbGJAskCPCj3h",
  "BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
  "6U91aKa8pmMxkJwBCfPTmUEfZi6dHe7DcFq2ALvB2tbB",
  "4xDsmeTWPNjgSVSS1VTfzFq3iHZhp77ffPkAmkZkdu71",
  "CapuXNQoDviLvU1PxFiizLgPNQCxrsag1uMeyk6zLVps",
  "9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
  "6LXutJvKUw8Q5ue2gCgKHQdAN4suWW8awzFVC6XCguFx",
  "HFqp6ErWHY6Uzhj8rFyjYuDya2mXUpYEk8VW75K9PSiY",
  "DSN3j1ykL3obAVNv7ZX49VsFCPe4LqzxHnmtLiPwY6xg",
  "69yhtoJR4JYPPABZcSNkzuqbaFbwHsCkja1sP1Q2aVT5",
  "HU23r7UoZbqTUuh3vA7emAGztFtqwTeVips789vqxxBw",
  "3LoAYHuSd7Gh8d7RTFnhvYtiTiefdZ5ByamU42vkzd76",
  "3CgvbiM3op4vjrrjH2zcrQUwsqh5veNVRjFCB9N6sRoD",
  "GP8StUXNYSZjPikyRsvkTbvRV1GBxMErb59cpeCJnDf1",
  "7iWnBRRhBCiNXXPhqiGzvvBkKrvFSWqqmxRyu9VyYBxE",
];

const tipIx = SystemProgram.transfer({
  fromPubkey: signer.publicKey,
  toPubkey: new PublicKey(
    TIP_ACCOUNTS[Math.floor(Math.random() * TIP_ACCOUNTS.length)]
  ),
  lamports: 1_000_000, // 0.001 SOL minimum
});

// Add tipIx to your transaction, sign, then send via tx.jup.ag
```

```typescript @solana/kit expandable theme={null}
import { getTransferSolInstruction } from "@solana-program/system";

const TIP_ACCOUNTS = [
  "GGztQqQ6pCPaJQnNpXBgELr5cs3WwDakRbh1iEMzjgSJ",
  "2MFoS3MPtvyQ4Wh4M9pdfPjz6UhVoNbFbGJAskCPCj3h",
  "BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
  "6U91aKa8pmMxkJwBCfPTmUEfZi6dHe7DcFq2ALvB2tbB",
  "4xDsmeTWPNjgSVSS1VTfzFq3iHZhp77ffPkAmkZkdu71",
  "CapuXNQoDviLvU1PxFiizLgPNQCxrsag1uMeyk6zLVps",
  "9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
  "6LXutJvKUw8Q5ue2gCgKHQdAN4suWW8awzFVC6XCguFx",
  "HFqp6ErWHY6Uzhj8rFyjYuDya2mXUpYEk8VW75K9PSiY",
  "DSN3j1ykL3obAVNv7ZX49VsFCPe4LqzxHnmtLiPwY6xg",
  "69yhtoJR4JYPPABZcSNkzuqbaFbwHsCkja1sP1Q2aVT5",
  "HU23r7UoZbqTUuh3vA7emAGztFtqwTeVips789vqxxBw",
  "3LoAYHuSd7Gh8d7RTFnhvYtiTiefdZ5ByamU42vkzd76",
  "3CgvbiM3op4vjrrjH2zcrQUwsqh5veNVRjFCB9N6sRoD",
  "GP8StUXNYSZjPikyRsvkTbvRV1GBxMErb59cpeCJnDf1",
  "7iWnBRRhBCiNXXPhqiGzvvBkKrvFSWqqmxRyu9VyYBxE",
];

const tipIx = getTransferSolInstruction({
  source: signer,
  destination: address(
    TIP_ACCOUNTS[Math.floor(Math.random() * TIP_ACCOUNTS.length)]
  ),
  amount: 1_000_000n, // 0.001 SOL minimum
});

// Add tipIx to your transaction, sign, then send via tx.jup.ag
```

## Best practices

* **Randomise tip accounts:** there are 16 tip receiver accounts. Randomise which one you send to on each transaction to reduce write-lock contention across concurrent submissions.
* **Run in parallel with your own RPC:** submitting the same signed transaction through both `tx.jup.ag` and your existing RPC is a zero-risk way to compare landing. Whichever confirms first wins; the duplicate is dropped by the network.
* **Poll for confirmation on your own RPC:** `tx.jup.ag` is send-only. Confirm with the blockhash and `lastValidBlockHeight` from your transaction. If the blockhash expires without confirmation, rebuild with a fresh blockhash and resubmit.
* **Simulate on your own RPC first:** `tx.jup.ag` does not simulate, so run any preflight simulation against your own RPC before submitting.
* **Colocate for lowest latency:** Jupiter's API gateway runs in [six AWS regions](https://developers.jup.ag/docs/portal/latency.md). Deploy your servers in the nearest region to minimise round-trip time.

## Why it lands fast

Jupiter's transaction landing stack is purpose-built for high throughput and low latency:

* **Direct to Beam.** `tx.jup.ag` routes straight to the landing infrastructure. The legacy `api.jup.ag/tx/v1/submit` proxies through the API gateway first, so `tx.jup.ag` removes that hop.
* **[SWQoS via high-stake validator](https://solana.com/developers/guides/advanced/stake-weighted-qos).** Jupiter operates [one of the highest-staked validators on Solana](https://solanabeach.io/validators). Solana's Stake-Weighted Quality of Service (SWQoS) reserves \~80% of a leader's TPU capacity for staked validators proportional to their stake. Higher stake means more reserved bandwidth when forwarding transactions to the current leader.
* **Beam (custom TPU forwarder).** Jupiter's own TPU client bypasses standard RPC nodes and sends transactions directly to leaders via staked QUIC connections. This removes intermediaries that could be malicious actors and eliminates RPC processing overhead.
* **[DoubleZero](https://doublezero.xyz/).** Dedicated fiber network for validator communication, with FPGA-based spam filtering and transaction deduplication at the network edge. Cleaner transaction set, faster propagation between validators.
* **Ultra-low-latency fiber.** Private fiber links from Tokyo to Frankfurt reduce physical propagation time between Jupiter's infrastructure and leader validators globally.

## How this differs from /execute

| | `/execute` | `tx.jup.ag` |
| - | - | - |
| **Paired with** | `/order` (meta-aggregator flow) | `/build` or any signed transaction |
| **Interface** | REST `POST /swap/v2/execute` | Solana JSON-RPC `sendTransaction` |
| **Fee model** | Jupiter swap fees (included in the order) | SOL tips (minimum 0.001 SOL) |
| **Transaction source** | Must match the exact transaction from `/order` | Any valid signed Solana transaction |
| **Use case** | Managed swap execution for `/order` transactions | Any transaction: `/build` swaps, non-Jupiter transactions |

## API reference

**Endpoint:** `POST https://tx.jup.ag`, JSON-RPC method `sendTransaction`.

See the [transaction submission API reference](https://developers.jup.ag/docs/api-reference/transaction/submit.md) for the request and response shape.

> [!NOTE]
>
> The legacy REST endpoint `POST https://api.jup.ag/tx/v1/submit` still accepts the same signed transactions, but `tx.jup.ag` is the recommended path going forward.

## Related

* [Build](https://developers.jup.ag/docs/swap/build.md): build custom transactions with `/build` (use `tipAmount` for automatic tip inclusion)
* [Order & Execute](https://developers.jup.ag/docs/swap/order-and-execute.md): the standard swap flow using `/order` + `/execute`
* [Reduce Latency](https://developers.jup.ag/docs/swap/advanced/reduce-latency.md): use `mode=fast` on `/build` for faster routing
* [Latency and Server Locations](https://developers.jup.ag/docs/portal/latency.md): API gateway regions and colocation guidance


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.