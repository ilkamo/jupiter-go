> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Cancel a DCA Order

> Withdraw the unfilled remainder of a DCA order through the two-step cancel flow.

DCA orders cannot be edited, only cancelled. Cancelling returns the unfilled remainder to your wallet; rounds that have already filled are not reversed. Every request needs the JWT from [authentication](https://developers.jup.ag/docs/trigger/authentication.md).

> [!NOTE]
>
> There is no update endpoint (`PATCH /trigger/v2/orders/dca/{id}` returns `404`), and no pause or resume: cancelling is one-way. To change, restart, or resume a DCA, cancel it and [create a new one](https://developers.jup.ag/docs/trigger/dca.md).

## Cancelling an order

Cancelling uses the same two-step withdrawal flow as price orders.

**Step 1** moves the order to `withdrawing` and returns an unsigned withdrawal transaction, along with `refundAmount` and `roundsRemaining`:

```typescript theme={null}
const cancel = await fetch(`${BASE}/orders/dca/cancel/${orderId}`, { method: "POST", headers }).then((r) => r.json());
// { id, roundsRemaining, refundAmount, transaction, requestId }
```

**Step 2** signs that transaction and confirms the withdrawal:

```typescript theme={null}
const tx = VersionedTransaction.deserialize(Buffer.from(cancel.transaction, "base64"));
tx.sign([wallet]);

const confirmed = await fetch(`${BASE}/orders/dca/confirm-cancel/${orderId}`, {
  method: "POST",
  headers,
  body: JSON.stringify({
    signedTransaction: Buffer.from(tx.serialize()).toString("base64"),
    cancelRequestId: cancel.requestId,
  }),
}).then((r) => r.json());
// { id, txSignature }
```

If the withdrawal does not land, call `cancel` again to re-craft it; the order stays in `withdrawing` and the request is idempotent. Cancelling is one-way: once initiated, the order stops filling and cannot return to `active` (a failed `confirm-cancel` leaves it in `withdrawing` so you can retry, it does not roll back). If you change your mind, complete the withdrawal to reclaim the unfilled funds, then [create a new order](https://developers.jup.ag/docs/trigger/dca.md) to resume.

## Errors

| Status | Meaning |
| :- | :- |
| `400` | Order not in a cancellable state (`Cannot cancel DCA order in '<state>' state`), or a stale `cancelRequestId` (`Withdrawal transaction not found in cache`) |
| `401` | Missing or expired JWT |
| `404` | `DCA order not found or you do not own it` |

See the [DCA errors](https://developers.jup.ag/docs/trigger/dca.md#errors) table for the full list.

## Related

> [card] [Create a DCA Order](https://developers.jup.ag/docs/trigger/dca.md)
>
> Set up the deposit and schedule.

> [card] [Track a DCA Order](https://developers.jup.ag/docs/trigger/dca-history.md)
>
> Monitor progress, fills, and state.

> [card] [Cancel DCA (API)](https://developers.jup.ag/docs/api-reference/trigger/cancel-dca.md)
>
> Full cancel request and response schema.

> [card] [Authentication](https://developers.jup.ag/docs/trigger/authentication.md)
>
> The challenge-response JWT flow every request needs.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.