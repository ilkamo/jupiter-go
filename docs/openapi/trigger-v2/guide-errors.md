> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Error Reference

> HTTP status codes returned by each Trigger Order API endpoint and what causes them.

Most error responses are JSON with an `error` message:

```json theme={null}
{ "error": "Order not found" }
```

Validation failures add a `details` object with per-field messages, so parse `details` when you get a `400`:

```json theme={null}
{ "error": "Request validation failed", "details": { "orderSubType": "Invalid input" } }
```

`GET /vault/register` also returns the existing vault in `details` on a `409`. Access errors from the API gateway (for example, an API key without Trigger access) use a different shape: `{ "code": 403, "message": "..." }`.

This page covers the price-order and shared (auth, vault, deposit) endpoints. For DCA endpoint errors (`/orders/dca`, `/orders/history/dca`, `/orders/dca/cancel`, `/confirm-cancel`), see [DCA errors](https://developers.jup.ag/docs/trigger/dca.md#errors).

## Authentication errors apply everywhere

Every endpoint except `POST /auth/challenge` and `POST /auth/verify` requires a JWT in the `Authorization: Bearer <token>` header. A missing, malformed, or expired token returns:

```json theme={null}
{ "error": "Unauthorized" }
```

with status `401`. When you see a `401` on an authenticated call, re-run the [challenge-response flow](https://developers.jup.ag/docs/trigger/authentication.md) to get a fresh token. The tables below list the errors specific to each endpoint on top of this shared `401`. The DCA endpoints share this `401` rule; their endpoint-specific errors are on the [DCA pages](https://developers.jup.ag/docs/trigger/dca.md#errors).

## Authentication

### POST /auth/challenge

| Status | Meaning |
| :- | :- |
| `400` | Invalid request body (bad `walletPubkey`, or `type` is not `message` or `transaction`) |
| `500` | Failed to generate the challenge |

### POST /auth/verify

| Status | Meaning |
| :- | :- |
| `400` | Challenge not found, expired, or invalid request body |
| `401` | Invalid signature, or the challenge has expired |
| `500` | Internal server error |

## Vault

### GET /vault

| Status | Meaning |
| :- | :- |
| `404` | Vault not found. Call `GET /vault/register` to create one |
| `500` | Failed to get vault info |

### GET /vault/register

| Status | Meaning |
| :- | :- |
| `409` | Vault already registered. The existing vault is returned in `details` |
| `500` | Failed to generate or store the vault |

## Deposit

### POST /deposit/craft

| Status | Meaning |
| :- | :- |
| `400` | Validation error, or `userAddress` equals the vault address |
| `403` | `userAddress` does not match the authenticated wallet, or no vault is registered for this user |
| `500` | Failed to fetch vault info or craft the deposit transaction |

## Orders

### POST /orders/price

| Status | Meaning |
| :- | :- |
| `400` | Validation error (price relationships, slippage range, expiry), no vault registered, `userPubkey` equals the vault address, an invalid signed deposit transaction, or a duplicate deposit |
| `403` | `userPubkey` does not match the authenticated wallet |
| `500` | Failed to fetch the trigger mint price or create the order |

### PATCH /orders/price/{id}

| Status | Meaning |
| :- | :- |
| `400` | Validation error, or the order is not in the `open` state (only `open` orders can be updated) |
| `404` | Order not found |
| `500` | Failed to fetch or update the order |

### POST /orders/price/cancel/{id}

| Status | Meaning |
| :- | :- |
| `400` | The order is not yet marked ready to cancel. Retry shortly |
| `403` | The order belongs to a different wallet |
| `404` | Order not found |
| `500` | Failed to initiate the cancellation or craft the withdrawal transaction |

### POST /orders/price/confirm-cancel/{id}

| Status | Meaning |
| :- | :- |
| `400` | Invalid transaction format, an invalid or mismatched withdrawal transaction, or a withdrawal validation error |
| `404` | Order not found |
| `500` | Failed to prepare or confirm the withdrawal |

## History

### GET /orders/history

| Status | Meaning |
| :- | :- |
| `403` | You do not have permission to view these orders |
| `404` | Order not found (when requesting a single order by id) |

## Related

> [card] [DCA Errors](https://developers.jup.ag/docs/trigger/dca.md#errors)
>
> Errors for the DCA create, track, and cancel endpoints.

> [card] [Integration Flow](https://developers.jup.ag/docs/trigger/lifecycle.md)
>
> The full call sequence and order state machine.

> [card] [Manage Orders](https://developers.jup.ag/docs/trigger/manage-orders.md)
>
> Update, cancel, and recover orders.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.