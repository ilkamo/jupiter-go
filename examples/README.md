# New client examples

The root package is the v1 implementation in development. See
[development status](../docs/development-status.md) for current coverage.

The current default collects no maintainer fee. Jupiter and network fees can
still apply. `WithReferral(account, 50)` configures a 0.5% managed-order referral;
provision the referral and applicable token accounts separately. Always inspect
returned fee fields before signing. Missing referral token accounts can prevent
referral collection.

Run a quote without signing or sending a transaction:

```sh
JUPITER_API_KEY=your-key go run ./examples/quote
```

A key may be omitted where Jupiter permits keyless access. The example requests
0.1 SOL in raw units and prints the USDC output amount in raw units. It omits a
`taker`, so the response is a quote without a signable transaction.

For managed execution, request an order with a taker, review its fees and output,
partially sign the returned transaction with an appropriate Solana signing
library, then call `client.Swap.Execute` with the same request ID and signed
base64 transaction. Preserve the message, blockhash, and existing signatures.
The SDK does not automatically re-create or retry an ambiguous execution.

For custom transactions, `client.Swap.Build` returns raw instructions. Assemble,
simulate, sign, and submit them through your own pipeline; these transactions
cannot be passed to `Swap.Execute`.
