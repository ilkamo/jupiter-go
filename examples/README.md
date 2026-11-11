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

Every Swap result exposes `Metadata.StatusCode`, `Metadata.RequestID`, and
`Metadata.Headers`. Headers include upstream rate limits and `Retry-After`.
The HTTP request ID is separate from `OrderResponse.RequestID`, which belongs to
order execution. Program labels are available as `result.Labels` alongside the
same metadata. Always check the error before using response data; a non-nil
response may accompany an error so its metadata remains inspectable.

Optional scalar response fields use pointers where zero is meaningful. Check for
nil before reading `Gasless`, execution `Code`, build `TransactionVersion`, or
`BlockhashWithMetadata`. A missing execution code does not confirm success.
`ExecutionError.Code` is nil when the upstream response supplied no code.
