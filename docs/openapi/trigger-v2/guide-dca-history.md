> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Track a DCA Order

> Monitor a DCA order's progress, fills, and state through the history endpoints.

Track a DCA order's schedule, progress, and fill history through the order-history endpoints. Every request needs the JWT from [authentication](https://developers.jup.ag/docs/trigger/authentication.md).

## Tracking an order

List your DCA orders, or fetch one by ID. Both require the JWT.

```typescript theme={null}
const list = await fetch(`${BASE}/orders/history/dca?state=active&limit=20`, { headers }).then((r) => r.json());
const detail = await fetch(`${BASE}/orders/history/dca/${orderId}`, { headers }).then((r) => r.json());
```

> [!NOTE]
>
> The per-order detail path requires the `dca` segment: `GET /trigger/v2/orders/history/dca/{id}`.

List query parameters: `state` (`active` or `past`), `mint` (matches input or output), `limit` (1–100, default 20), `offset`, `sort` (`updated_at`, `created_at`, `next_fill_at`), and `dir` (`asc` / `desc`).

Each order reports its schedule, progress, and event history:

| Field | Description |
| :- | :- |
| `inputAmountInitial` / `inputAmountRemaining` | Total deposited, and amount still in the vault. |
| `amountPerRound` | Input swapped each round (the last round absorbs any dust). |
| `outputAmountTotal` / `inputAmountUsed` | Output received and input consumed across filled rounds. |
| `numberOfRounds` / `roundsFilled` / `fillPercent` | Total rounds, rounds completed, and progress from 0 to 1. |
| `nextFillAt` / `lastFillAt` | Next scheduled round, and the most recent fill. |
| `state` / `displayState` | Raw and human-friendly status (see below). |
| `jlEnabled` / `jlYieldUsd` | Whether [Earn While You Wait](https://developers.jup.ag/docs/trigger/dca.md#earn-while-you-wait) is on, and the yield accrued so far in USD (`null` when not JL-enabled or a required price is uncached). |
| `events` | Ordered history of `deposit`, `fill`, `withdrawal`, and `cancelled` events. |

> [accordion] Full order response
>
> ```json theme={null}
> {
>   "id": "019f02b5-0400-72cc-8943-007c4310a3de",
>   "requestId": "2ea435dc-1068-466c-8fff-c9ad74d076e8",
>   "userPubkey": "8gs8rXavMQRqoAoMhCfo79r2wndgSn93FLSWErdHp82n",
>   "vaultPubkey": "66xK9AE6oFpGAyn9eCWE3LtYwUZKkeUxbfjFSNmFnkBo",
>   "inputMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
>   "outputMint": "So11111111111111111111111111111111111111112",
>   "inputAmountInitial": "20000000",
>   "inputAmountRemaining": "20000000",
>   "amountPerRound": "10000000",
>   "outputAmountTotal": "0",
>   "inputAmountUsed": "0",
>   "orderType": "time_based",
>   "minPriceUsd": null,
>   "maxPriceUsd": null,
>   "triggerMint": null,
>   "retryWindowSeconds": 30,
>   "jlEnabled": false,
>   "jlYieldUsd": null,
>   "numberOfRounds": 2,
>   "intervalSeconds": 60,
>   "beginFillAt": "2026-06-26T07:53:57.767Z",
>   "nextFillAt": "2026-06-26T07:53:57.767Z",
>   "lastFillAt": null,
>   "roundsFilled": 0,
>   "fillPercent": 0,
>   "state": "active",
>   "displayState": "active",
>   "createdAt": "2026-06-26T06:54:05.327Z",
>   "updatedAt": "2026-06-26T06:54:05.924Z",
>   "events": [
>     {
>       "type": "deposit",
>       "timestamp": "2026-06-26T06:54:05.924Z",
>       "txSignature": "4R18eTrkDvmZFPq87pznBPpVCpk1G2FQapWVks4ELLNYxAQZdwoDaxnnAHN2ZvBFrRvbnJSAPTnhutgj8ZzLFwgV",
>       "inputAmount": "20000000",
>       "state": "success"
>     }
>   ]
> }
> ```
>
> Reconcile against `state` and `events`, not `inputAmountRemaining`, which is not zeroed after a cancellation.

### Order states

The response carries two status fields: `state` is the raw internal status, and `displayState` is the human-friendly version most integrations show.

| `state` | `displayState` | Meaning |
| :- | :- | :- |
| `depositing` | `pending` | Deposit in progress. |
| `active` | `active` | Scheduled and waiting for the next round. |
| `executing` | `executing` | A round is currently filling. |
| `withdrawing` | `pending_withdraw` | A cancellation is in progress. |
| `completed` | `completed` | All rounds filled. |
| `cancelled` | `cancelled` | Cancelled by you. |
| `deposit_failed` | `failed` | The deposit failed on-chain. |

For how these states fit the end-to-end flow, see the [Integration Flow](https://developers.jup.ag/docs/trigger/lifecycle.md#order-state-lifecycle).

## How rounds fill

When a round comes due, the keeper swaps `amountPerRound` and **delivers the output straight to your wallet**, and pays the swap's transaction fee itself. You never withdraw filled output manually; only the unfilled remainder is returned, and only when you [cancel](https://developers.jup.ag/docs/trigger/dca-cancel.md). A round that is due now typically fills within a few seconds.

Each fill appends an event and advances `roundsFilled`, `outputAmountTotal`, `inputAmountUsed`, and `nextFillAt`:

```json theme={null}
{
  "type": "fill",
  "timestamp": "2026-06-26T18:42:11.469Z",
  "txSignature": "HuNMpjFkTjVrCYMN6gTPWV9ay3GSxTxM7sZj4WNzjvzQoe4J5CR71MnTpuMKH9rtkDoDCTWu6rYRM12oR9cpbb4",
  "roundNumber": 1,
  "inputAmount": "10000000",
  "outputAmount": "137589159",
  "state": "success"
}
```

A `price_conditional` round that is out of band does not fill: the order stays `active` with `roundsFilled` unchanged until the price re-enters the range, then resumes on the next interval.

If a round cannot execute (for example, not enough liquidity within slippage), the keeper retries it within a window of `min(max(30, intervalSeconds / 2), 7200)` seconds, then reschedules it to the next interval. `roundsFilled` stays the same and `nextFillAt` advances. A round's `fill` event carries a `state` of `success`, `failed`, `rescheduled`, or `pending`.

## Fees and slippage

Each round executes as a standard Jupiter swap, run by the keeper. Two things follow from that:

* **Slippage is automatic and not configurable.** DCA orders have no slippage parameter. The keeper sets slippage per round from live market conditions at fill time; you cannot set or cap it.
* **The swap fee is included in the output.** Any swap fee is already reflected in the `outputAmount` you receive each round. Trigger V2 does not support integrator fees.

The keeper pays the network and priority fees for each fill, and the output settles directly to your wallet.

## Related

> [card] [Create a DCA Order](https://developers.jup.ag/docs/trigger/dca.md)
>
> Set up the deposit and schedule.

> [card] [Cancel a DCA Order](https://developers.jup.ag/docs/trigger/dca-cancel.md)
>
> Withdraw the unfilled remainder.

> [card] [DCA History (API)](https://developers.jup.ag/docs/api-reference/trigger/dca-history.md)
>
> Full history request and response schema.

> [card] [Authentication](https://developers.jup.ag/docs/trigger/authentication.md)
>
> The challenge-response JWT flow every request needs.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.