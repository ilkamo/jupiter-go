> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Order (Trigger API)

> Create a new trigger order



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/price
openapi: 3.0.0
info:
  title: Jupiter Trigger Order API V2
  version: 2.0.0
  description: >-
    Jupiter Trigger Order API V2 with vault-based deposits, JWT authentication,
    and advanced order types.
servers:
  - url: https://api.jup.ag/trigger/v2
    description: Jupiter Trigger Order API V2 Endpoint
security: []
paths:
  /orders/price:
    post:
      summary: Create price order
      description: |
        Create a new price order. Supports three order types:
        - `single`: triggers when price crosses above/below threshold
        - `oco`: one-cancels-other (take-profit + stop-loss pair)
        - `otoco`: parent trigger activates a TP/SL pair on fill
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                orderType:
                  type: string
                  enum:
                    - single
                    - oco
                    - otoco
                depositRequestId:
                  type: string
                  description: From /deposit/craft response
                depositSignedTx:
                  type: string
                  description: Base64-encoded signed deposit transaction
                userPubkey:
                  type: string
                inputMint:
                  type: string
                inputAmount:
                  type: string
                  description: Amount in smallest unit
                outputMint:
                  type: string
                triggerMint:
                  type: string
                  description: Token mint to monitor for price
                triggerCondition:
                  type: string
                  enum:
                    - above
                    - below
                  description: Single/OTOCO only
                triggerPriceUsd:
                  type: number
                  description: >-
                    Single/OTOCO trigger price. For single orders, required
                    unless trailingBps is set.
                trailingBps:
                  type: number
                  minimum: 50
                  maximum: 9000
                  description: >-
                    Trailing stop loss trail distance in basis points (50-9000).
                    Single orders only; set instead of triggerPriceUsd.
                    Direction from triggerCondition: below = sell-below
                    (triggerMint === inputMint), above = buy-above (triggerMint
                    === outputMint).
                slippageBps:
                  type: number
                  description: Slippage tolerance (0-10000 bps)
                tpPriceUsd:
                  type: number
                  description: Take-profit price (OCO/OTOCO)
                slPriceUsd:
                  type: number
                  description: Stop-loss price (OCO/OTOCO)
                tpSlippageBps:
                  type: number
                  description: Take-profit slippage (OCO/OTOCO)
                slSlippageBps:
                  type: number
                  description: Stop-loss slippage (OCO/OTOCO)
                expiresAt:
                  type: number
                  description: Expiration timestamp (milliseconds)
              required:
                - orderType
                - depositRequestId
                - depositSignedTx
                - userPubkey
                - inputMint
                - inputAmount
                - outputMint
                - triggerMint
                - expiresAt
            example:
              orderType: single
              depositRequestId: 01234567-89ab-cdef-0123-456789abcdef
              depositSignedTx: AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
              userPubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              inputMint: So11111111111111111111111111111111111111112
              inputAmount: '1000000000'
              outputMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
              triggerMint: So11111111111111111111111111111111111111112
              triggerCondition: above
              triggerPriceUsd: 200
              slippageBps: 100
              expiresAt: 1735689600000
      responses:
        '200':
          description: Order created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrderResponse'
              example:
                id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
                txSignature: 5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d...
                depositConfirmed: true
        '400':
          description: Validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Unauthorized
        '403':
          description: Forbidden
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    OrderResponse:
      type: object
      description: >-
        Response for a price order create. The deposit lands synchronously
        during the call, so depositConfirmed is returned. DCA create and
        confirm-cancel use TxSignatureResponse instead.
      properties:
        id:
          type: string
          description: Order UUID
        txSignature:
          type: string
          description: On-chain signature of the deposit transaction
        depositConfirmed:
          type: boolean
          description: Whether the deposit has landed and the order is active
        message:
          type: string
          description: >-
            Additional context about the deposit state, present on idempotent
            retries
      required:
        - id
        - txSignature
        - depositConfirmed
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        details:
          type: object
          additionalProperties:
            type: string
      required:
        - error
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT token from the challenge-response auth flow

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.