# Jupiter API endpoint inventory

Reviewed 2026-11-11 on branch `v1`. This inventories all 82 API-reference pages
linked from the official [documentation index](https://developers.jup.ag/docs/llms.txt).
Every linked reference was fetched successfully and its OpenAPI operation header
and server inspected. This is a discovery and contract inventory, not a claim
that every operation is implemented or that its full schema has been audited.

## Provenance and interpretation

[`api-reference-manifest.json`](api-reference-manifest.json) records the source URL,
OpenAPI schema identifier, operation, server, security declarations, pagination
parameter names, implementation status, and SHA-256 of each retrieved Markdown
reference. Hashes identify the reviewed content; only the existing three Swap
snapshots are vendored. Fetch and compare the reference hash before implementing
an operation; review changed contracts rather than silently accepting drift.

Authentication below reports the published security declarations, not a verified
live authorization result. Gateway documentation permits keyless access for some
products while individual specs still declare API-key security. Keep API keys
optional at the transport level and verify each product's effective requirements.
Bearer credentials must belong to a wallet session, not a shared global client.

Pagination lists candidate query controls from each operation reference. A limit
alone does not imply that the result supports fetching successive pages. `—`
means no named pagination control was found, not that results are unlimited.

No write-operation replay guarantee has been established by this inventory.
Every operation declares its own retry policy in the transport request. Swap
order, build, and program labels retry transient failures; Swap execute makes one
attempt. Operations in other products must be reviewed before choosing a policy;
state-changing GETs must not inherit a read-only policy.

## Coverage

- swap: 4 operations; 4 methods implemented, response parity review pending.
- tokens: 7 operations; 4 discovery operations implemented, 3 verification operations pending.
- price: 1 operation; implemented.
- transaction: 1 operation; not implemented.
- lend: 15 operations; not implemented.
- trigger: 18 operations; not implemented.
- prediction: 27 operations; not implemented.
- send: 4 operations; not implemented.
- studio: 5 operations; not implemented.

## Operations

`Partial` means a public method exists but the plan's full contract acceptance
criteria are still outstanding. `Pending` means no new SDK method exists.

| Product | Method and URL | Declared security | Pagination controls | SDK / tests | Reference |
| --- | --- | --- | --- | --- | --- |
| swap | `GET https://api.jup.ag/swap/v2/order` | ApiKeyAuth | — | Partial / request and response contract tests | [Get Order (Swap API)](https://developers.jup.ag/docs/api-reference/swap/order.md) |
| swap | `POST https://api.jup.ag/swap/v2/execute` | ApiKeyAuth | — | Partial / request and response contract tests | [Execute Transaction](https://developers.jup.ag/docs/api-reference/swap/execute.md) |
| swap | `GET https://api.jup.ag/swap/v2/build` | ApiKeyAuth | — | Partial / request and response contract tests | [Build Transaction](https://developers.jup.ag/docs/api-reference/swap/build.md) |
| swap | `GET https://api.jup.ag/swap/v2/program-id-to-label` | ApiKeyAuth | — | Partial / success metadata and redirect tests | [Program ID to Label](https://developers.jup.ag/docs/api-reference/swap/program-id-to-label.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/search` | ApiKeyAuth | — | Implemented / request and response contract tests | [Search](https://developers.jup.ag/docs/api-reference/tokens/search.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/tag` | ApiKeyAuth | — | Implemented / request and response contract tests | [Tag](https://developers.jup.ag/docs/api-reference/tokens/tag.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/{category}/{interval}` | ApiKeyAuth | limit | Implemented / request and response contract tests | [Category](https://developers.jup.ag/docs/api-reference/tokens/category.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/recent` | ApiKeyAuth | — | Implemented / request and response contract tests | [Recent](https://developers.jup.ag/docs/api-reference/tokens/recent.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/verify/express/check-eligibility` | ApiKeyAuth | — | Pending / none | [Check Eligibility](https://developers.jup.ag/docs/api-reference/tokens/verify-check-eligibility.md) |
| tokens | `GET https://api.jup.ag/tokens/v2/verify/express/craft-txn` | ApiKeyAuth | — | Pending / none | [Craft Transaction](https://developers.jup.ag/docs/api-reference/tokens/verify-craft-txn.md) |
| tokens | `POST https://api.jup.ag/tokens/v2/verify/express/execute` | ApiKeyAuth | — | Pending / none | [Execute Verification (Tokens API)](https://developers.jup.ag/docs/api-reference/tokens/verify-execute.md) |
| price | `GET https://api.jup.ag/price/v3/` | ApiKeyAuth | — | Implemented / request and response contract tests | [Get Price (Price API V3)](https://developers.jup.ag/docs/api-reference/price/index.md) |
| transaction | `POST https://tx.jup.ag/` | ApiKeyAuth | — | Pending / none | [Send Transaction](https://developers.jup.ag/docs/api-reference/transaction/submit.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/deposit` | ApiKeyAuth | — | Pending / none | [Deposit](https://developers.jup.ag/docs/api-reference/lend/earn/deposit.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/withdraw` | ApiKeyAuth | — | Pending / none | [Withdraw](https://developers.jup.ag/docs/api-reference/lend/earn/withdraw.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/mint` | ApiKeyAuth | — | Pending / none | [Mint](https://developers.jup.ag/docs/api-reference/lend/earn/mint.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/redeem` | ApiKeyAuth | — | Pending / none | [Redeem](https://developers.jup.ag/docs/api-reference/lend/earn/redeem.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/deposit-instructions` | ApiKeyAuth | — | Pending / none | [Deposit Instructions](https://developers.jup.ag/docs/api-reference/lend/earn/deposit-instructions.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/withdraw-instructions` | ApiKeyAuth | — | Pending / none | [Withdraw Instructions](https://developers.jup.ag/docs/api-reference/lend/earn/withdraw-instructions.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/mint-instructions` | ApiKeyAuth | — | Pending / none | [Mint Instructions](https://developers.jup.ag/docs/api-reference/lend/earn/mint-instructions.md) |
| lend | `POST https://api.jup.ag/lend/v1/earn/redeem-instructions` | ApiKeyAuth | — | Pending / none | [Redeem Instructions](https://developers.jup.ag/docs/api-reference/lend/earn/redeem-instructions.md) |
| lend | `GET https://api.jup.ag/lend/v1/earn/tokens` | ApiKeyAuth | — | Pending / none | [Tokens](https://developers.jup.ag/docs/api-reference/lend/earn/tokens.md) |
| lend | `GET https://api.jup.ag/lend/v1/earn/positions` | ApiKeyAuth | — | Pending / none | [Earn Positions](https://developers.jup.ag/docs/api-reference/lend/earn/positions.md) |
| lend | `GET https://api.jup.ag/lend/v1/earn/earnings` | ApiKeyAuth | — | Pending / none | [Earnings](https://developers.jup.ag/docs/api-reference/lend/earn/earnings.md) |
| lend | `GET https://api.jup.ag/lend/v1/borrow/vaults` | ApiKeyAuth | — | Pending / none | [Vaults](https://developers.jup.ag/docs/api-reference/lend/borrow/vaults.md) |
| lend | `GET https://api.jup.ag/lend/v1/borrow/positions` | ApiKeyAuth | — | Pending / none | [Borrow Positions](https://developers.jup.ag/docs/api-reference/lend/borrow/positions.md) |
| lend | `POST https://api.jup.ag/lend/v1/borrow/operate` | ApiKeyAuth | — | Pending / none | [Operate](https://developers.jup.ag/docs/api-reference/lend/borrow/operate.md) |
| lend | `POST https://api.jup.ag/lend/v1/borrow/operate-instructions` | ApiKeyAuth | — | Pending / none | [Operate Instructions](https://developers.jup.ag/docs/api-reference/lend/borrow/operate-instructions.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/auth/challenge` | ApiKeyAuth | — | Pending / none | [Request Challenge](https://developers.jup.ag/docs/api-reference/trigger/challenge.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/auth/verify` | ApiKeyAuth | — | Pending / none | [Verify Challenge](https://developers.jup.ag/docs/api-reference/trigger/verify.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/auth/refresh` | ApiKeyAuth | — | Pending / none | [Refresh Token](https://developers.jup.ag/docs/api-reference/trigger/refresh.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/auth/logout` | ApiKeyAuth | — | Pending / none | [Log Out](https://developers.jup.ag/docs/api-reference/trigger/logout.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/auth/logout-all` | ApiKeyAuth, BearerAuth | — | Pending / none | [Log Out All](https://developers.jup.ag/docs/api-reference/trigger/logout-all.md) |
| trigger | `GET https://api.jup.ag/trigger/v2/vault` | ApiKeyAuth, BearerAuth | — | Pending / none | [Get Vault](https://developers.jup.ag/docs/api-reference/trigger/vault.md) |
| trigger | `GET https://api.jup.ag/trigger/v2/vault/register` | ApiKeyAuth, BearerAuth | — | Pending / none | [Register Vault](https://developers.jup.ag/docs/api-reference/trigger/vault-register.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/deposit/craft` | ApiKeyAuth, BearerAuth | — | Pending / none | [Craft Deposit](https://developers.jup.ag/docs/api-reference/trigger/deposit-craft.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/price` | ApiKeyAuth, BearerAuth | — | Pending / none | [Create Order (Trigger API)](https://developers.jup.ag/docs/api-reference/trigger/create-order.md) |
| trigger | `PATCH https://api.jup.ag/trigger/v2/orders/price/{orderId}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Update Order](https://developers.jup.ag/docs/api-reference/trigger/update-order.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/price/cancel/{orderId}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Cancel Order (Trigger API)](https://developers.jup.ag/docs/api-reference/trigger/cancel-order.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/price/confirm-cancel/{orderId}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Confirm Cancel](https://developers.jup.ag/docs/api-reference/trigger/confirm-cancel.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/dca` | ApiKeyAuth, BearerAuth | — | Pending / none | [Create DCA Order](https://developers.jup.ag/docs/api-reference/trigger/create-dca.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/dca/cancel/{id}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Cancel DCA Order](https://developers.jup.ag/docs/api-reference/trigger/cancel-dca.md) |
| trigger | `POST https://api.jup.ag/trigger/v2/orders/dca/confirm-cancel/{id}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Confirm Cancel DCA Order](https://developers.jup.ag/docs/api-reference/trigger/confirm-cancel-dca.md) |
| trigger | `GET https://api.jup.ag/trigger/v2/orders/history/dca` | ApiKeyAuth, BearerAuth | limit, offset | Pending / none | [List DCA Orders](https://developers.jup.ag/docs/api-reference/trigger/dca-history.md) |
| trigger | `GET https://api.jup.ag/trigger/v2/orders/history/dca/{id}` | ApiKeyAuth, BearerAuth | — | Pending / none | [Get DCA Order](https://developers.jup.ag/docs/api-reference/trigger/dca-order.md) |
| trigger | `GET https://api.jup.ag/trigger/v2/orders/history` | ApiKeyAuth, BearerAuth | limit, offset | Pending / none | [Order History (Trigger API)](https://developers.jup.ag/docs/api-reference/trigger/order-history.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events` | ApiKeyAuth | end, start | Pending / none | [Get Events](https://developers.jup.ag/docs/api-reference/prediction/get-events.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/search` | ApiKeyAuth | limit | Pending / none | [Search Events](https://developers.jup.ag/docs/api-reference/prediction/search-events.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/{eventId}` | ApiKeyAuth | — | Pending / none | [Get Event](https://developers.jup.ag/docs/api-reference/prediction/get-event.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/suggested/{pubkey}` | ApiKeyAuth | — | Pending / none | [Get Suggested Events](https://developers.jup.ag/docs/api-reference/prediction/get-suggested-events.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/scores` | ApiKeyAuth | — | Pending / none | [Get Live Scores](https://developers.jup.ag/docs/api-reference/prediction/get-event-scores.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/{eventId}/score` | ApiKeyAuth | — | Pending / none | [Get Event Score](https://developers.jup.ag/docs/api-reference/prediction/get-event-score.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/{eventId}/markets` | ApiKeyAuth | end, start | Pending / none | [Get Event Markets](https://developers.jup.ag/docs/api-reference/prediction/get-event-markets.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/events/{eventId}/markets/{marketId}` | ApiKeyAuth | — | Pending / none | [Get Event Market](https://developers.jup.ag/docs/api-reference/prediction/get-event-market.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/markets/{marketId}` | ApiKeyAuth | — | Pending / none | [Get Market](https://developers.jup.ag/docs/api-reference/prediction/get-market.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/orders` | ApiKeyAuth | end, start | Pending / none | [Get Orders](https://developers.jup.ag/docs/api-reference/prediction/get-orders.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/orders/{orderPubkey}` | ApiKeyAuth | — | Pending / none | [Get Order (Prediction API)](https://developers.jup.ag/docs/api-reference/prediction/get-order.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/orders/status/{orderPubkey}` | ApiKeyAuth | — | Pending / none | [Get Order Status](https://developers.jup.ag/docs/api-reference/prediction/get-order-status.md) |
| prediction | `POST https://api.jup.ag/prediction/v1/orders` | ApiKeyAuth | — | Pending / none | [Create Order (Prediction API)](https://developers.jup.ag/docs/api-reference/prediction/create-order.md) |
| prediction | `POST https://api.jup.ag/prediction/v1/execute` | ApiKeyAuth | — | Pending / none | [Execute Order (Prediction API)](https://developers.jup.ag/docs/api-reference/prediction/execute.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/positions` | ApiKeyAuth | end, start | Pending / none | [Get Positions (Prediction API)](https://developers.jup.ag/docs/api-reference/prediction/get-positions.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/positions/{positionPubkey}` | ApiKeyAuth | — | Pending / none | [Get Position](https://developers.jup.ag/docs/api-reference/prediction/get-position.md) |
| prediction | `DELETE https://api.jup.ag/prediction/v1/positions/{positionPubkey}` | ApiKeyAuth | — | Pending / none | [Close Position](https://developers.jup.ag/docs/api-reference/prediction/close-position.md) |
| prediction | `DELETE https://api.jup.ag/prediction/v1/positions` | ApiKeyAuth | — | Pending / none | [Close All Positions](https://developers.jup.ag/docs/api-reference/prediction/close-all-positions.md) |
| prediction | `POST https://api.jup.ag/prediction/v1/positions/{positionPubkey}/claim` | ApiKeyAuth | — | Pending / none | [Claim Position](https://developers.jup.ag/docs/api-reference/prediction/claim-position.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/history` | ApiKeyAuth | end, start | Pending / none | [Get History](https://developers.jup.ag/docs/api-reference/prediction/get-history.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/orderbook/{marketId}` | ApiKeyAuth | — | Pending / none | [Get Orderbook](https://developers.jup.ag/docs/api-reference/prediction/get-orderbook.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/trading-status` | ApiKeyAuth | — | Pending / none | [Get Trading Status](https://developers.jup.ag/docs/api-reference/prediction/get-trading-status.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/profiles/{ownerPubkey}` | ApiKeyAuth | — | Pending / none | [Get Profile](https://developers.jup.ag/docs/api-reference/prediction/get-profile.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/profiles/{ownerPubkey}/pnl-history` | ApiKeyAuth | — | Pending / none | [Get PnL History](https://developers.jup.ag/docs/api-reference/prediction/get-pnl-history.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/trades` | ApiKeyAuth | — | Pending / none | [Get Trades](https://developers.jup.ag/docs/api-reference/prediction/get-trades.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/leaderboards` | ApiKeyAuth | limit | Pending / none | [Get Leaderboards](https://developers.jup.ag/docs/api-reference/prediction/get-leaderboards.md) |
| prediction | `GET https://api.jup.ag/prediction/v1/vault-info` | ApiKeyAuth | — | Pending / none | [Get Vault Info](https://developers.jup.ag/docs/api-reference/prediction/get-vault-info.md) |
| send | `POST https://api.jup.ag/send/v1/craft-send` | ApiKeyAuth | — | Pending / none | [Craft Send](https://developers.jup.ag/docs/api-reference/send/craft-send.md) |
| send | `POST https://api.jup.ag/send/v1/craft-clawback` | ApiKeyAuth | — | Pending / none | [Craft Clawback](https://developers.jup.ag/docs/api-reference/send/craft-clawback.md) |
| send | `GET https://api.jup.ag/send/v1/pending-invites` | ApiKeyAuth | page | Pending / none | [Pending Invites](https://developers.jup.ag/docs/api-reference/send/pending-invites.md) |
| send | `GET https://api.jup.ag/send/v1/invite-history` | ApiKeyAuth | page | Pending / none | [Invite History](https://developers.jup.ag/docs/api-reference/send/invite-history.md) |
| studio | `POST https://api.jup.ag/studio/v1/dbc-pool/create-tx` | ApiKeyAuth | — | Pending / none | [DBC Pool Create TX](https://developers.jup.ag/docs/api-reference/studio/dbc-pool-create-tx.md) |
| studio | `POST https://api.jup.ag/studio/v1/dbc-pool/submit` | ApiKeyAuth | — | Pending / none | [DBC Pool Submit](https://developers.jup.ag/docs/api-reference/studio/dbc-pool-submit.md) |
| studio | `GET https://api.jup.ag/studio/v1/dbc-pool/addresses/{mint}` | ApiKeyAuth | — | Pending / none | [DBC Pool Addresses by Mint](https://developers.jup.ag/docs/api-reference/studio/dbc-pool-addresses-by-mint.md) |
| studio | `POST https://api.jup.ag/studio/v1/dbc/fee` | ApiKeyAuth | — | Pending / none | [DBC Fee](https://developers.jup.ag/docs/api-reference/studio/dbc-fee.md) |
| studio | `POST https://api.jup.ag/studio/v1/dbc/fee/create-tx` | ApiKeyAuth | — | Pending / none | [DBC Fee Create TX](https://developers.jup.ag/docs/api-reference/studio/dbc-fee-create-tx.md) |

## Structural decisions before more clients

- **Transaction submission:** the reference specifies JSON-RPC `sendTransaction`
  at `https://tx.jup.ag`, rather than a REST request under the shared API gateway.
  Introduce an explicit transaction-service endpoint and JSON-RPC envelope/error
  handling before implementing it. Do not overload the REST path or base URL.
- **Trigger authentication:** model challenge, verification, refresh rotation,
  logout, and wallet-authorized operations together. Vault registration is a GET
  with a side effect, so HTTP-method-based retries are insufficient.
- **Tokens verification:** discovery and verification share a product but have
  different lifecycles. Verification crafts and submits payment transactions;
  it must not inherit a read-only discovery retry policy.
- **Lend:** transaction-building and instruction-building operations must expose
  distinct response types. Borrow amounts require their documented sign and
  sentinel semantics; do not reuse the positive-only swap amount validator.
- **Prediction:** beta endpoints still need a deliberate SDK compatibility
  boundary. Model transaction creation, execution, and position management with
  their actual response semantics before advertising support.
- **Studio:** creation includes upload URLs and submission. Separate authenticated
  API requests from external upload requests to prevent credential forwarding.
- **Perps and Lock:** on-chain integrations remain outside this REST inventory.
  Older Trigger/Recurring APIs remain migration topics rather than new defaults.

## Contract review and implementation order

1. Finish Swap schema parity and success metadata, including optional-field presence semantics. The existing snapshot and public types are the comparison
   baseline, not proof of complete response coverage.
2. Tokens discovery and Price are implemented with exact numeric response types
   and explicit missing-price reporting. Add verification after its payment workflow.
3. Design wallet-scoped Trigger sessions before adding its order methods.
4. Add Lend, Prediction, Send, Studio, and transaction submission using the
   boundaries above. Confirm schema stability, nullability, response envelopes,
   pagination semantics, and replay guarantees per operation before coding.
5. Reconcile guide-only endpoints and newly added index entries before claiming
   complete Jupiter coverage. This inventory covers the linked API references,
   not undocumented or independently advertised services.

Validation: all 82 index reference URLs were fetched; operation headers and
server URLs were present for every page; reference URLs and method/URL pairs are
unique. Counts and security/pagination metadata were derived from the downloaded
references, not from guessed endpoint names. No authenticated API requests or
on-chain transactions were made.
