> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Trigger Best Practices

> Guidelines for error handling, token compatibility, and integration patterns across price orders and DCA in the Trigger Order V2 API.

> [!NOTE]
>
> The Trigger Order V2 API is in **beta** and under active development. Behaviour, endpoints, and response formats may change as we iterate based on integrator feedback. Pin to specific response schemas and handle unknown fields gracefully.

## Reference Implementation

The [jup.ag](https://jup.ag) frontend is the canonical reference for how to interact with the V2 API. When in doubt about edge cases, validation logic, or UX patterns, refer to how jup.ag handles it. This includes:

* Order creation flows and parameter validation
* Price and slippage presentation
* Error handling and retry logic
* Order state management and polling

## Order Expiry (Price Orders)

Every price order requires an `expiresAt` timestamp. Unlike V1 where orders could exist indefinitely, V2 enforces expiry on all price orders. This allows the system to prune stale orders, improving execution quality for active orders.

| Use case | Recommended expiry |
| :- | :- |
| Short-term trades | 1-7 days |
| Swing trades | 7-30 days |
| Long-duration positions | 30 days, with periodic renewal |

For long-lived orders, set a 30-day expiry and renew by updating the order before it expires via the [update endpoint](https://developers.jup.ag/docs/trigger/manage-orders.md#update-an-order). There is no "no expiry" option by design.

DCA orders have no `expiresAt`. A DCA runs until its budget is spent: its duration is `orderCount × intervalSeconds`, and an optional `beginFillAt` (default now, at most 30 days out) sets when the first round starts.

## Slippage (Price Orders)

Slippage settings apply to price orders. DCA orders have no slippage parameter: the keeper sets slippage per round from live market conditions at fill time.

| Order type | Default | Recommendation |
| :- | :- | :- |
| Take-profit / buy below | Auto (RTSE) | Use the default unless you have specific requirements |
| Stop-loss / buy above | 20% (2000 bps) | Keep high for execution certainty; tighten only if you accept missed fills |
| Parent (OTOCO) | Auto (RTSE) | Use the default |

Stop-loss orders use a higher default because execution certainty matters more than price precision when cutting losses. If you lower stop-loss slippage, the order may not fill during volatile conditions.

## Error Handling

* **Validation errors** (400) include a `details` object with per-field messages. Parse these to surface specific issues to users.
* **Authentication errors** (401) mean the JWT has expired. Re-authenticate with a new challenge-response flow rather than retrying the same token.
* **Minimum order size** is 10 USD: per price order, or per DCA round (`inputAmount / orderCount`). Enforce this client-side to avoid unnecessary API calls.
* Handle unknown fields in responses gracefully. As the API evolves during beta, new fields may be added to response objects.

## Order Summary

The API does not validate whether your trigger price makes economic sense. If a user sets a buy order above market price or a sell order below market price, the keeper will execute as instructed and the difference is lost.

Display a human-readable summary of what the order will do before the user confirms. On [jup.ag](https://jup.ag), the confirmation shows a plain-language description like:

> Buy SOL with 15 USDC when price goes below $76.37 (-10% from market). If triggered, automatically set TP for SOL at $84.01.

This helps users catch mistakes before submitting. For OCO and OTOCO orders, include both the trigger conditions and the child order parameters in the summary.

For DCA, summarise the schedule (number of rounds × interval), the per-round amount (`inputAmount / orderCount`, with the last round absorbing any remainder), and for `price_conditional` orders the `[minPriceUsd, maxPriceUsd]` band and which mint is the trigger.

## Token Compatibility

* Tokens with transfer-fee or transfer-hook extensions are rejected at deposit unless whitelisted
* Input and output mints must be different

## Order Lifecycle

Poll the history endpoint to track order state; the states differ by family.

**Price orders** ([`GET /orders/history`](https://developers.jup.ag/docs/trigger/order-history#order-states)):

| State | Meaning | Action |
| :- | :- | :- |
| `open` | Active, waiting for trigger | No action needed |
| `filled` | Fully executed | Show completion to user |
| `cancelled` | Cancelled by user | Confirm funds returned to vault |
| `expired` | Past `expiresAt` | Prompt user to withdraw funds or recreate |
| `failed` | Execution failed | Show error, suggest retry with adjusted parameters |

**DCA** ([`GET /orders/history/dca`](https://developers.jup.ag/docs/trigger/dca-history#order-states)) uses a different set: `pending`, `active`, `executing`, `completed` (all rounds filled), `pending_withdraw`, `cancelled`, and `failed`. There is no `open` or `expired` state; a DCA ends at `completed` when its budget is spent.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.