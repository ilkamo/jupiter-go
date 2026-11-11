> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Order (Prediction API)

> Request an unsigned transaction to create a new order



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml post /orders
openapi: 3.0.3
info:
  title: Jupiter Prediction Market API
  version: 1.0.0
  description: Jupiter Prediction Market API Schema
servers:
  - url: https://api.jup.ag/prediction/v1
    description: Jupiter Prediction Market API Endpoint
security:
  - ApiKeyAuth: []
tags:
  - name: Events
  - name: Markets
  - name: Orders
  - name: Positions
  - name: History
  - name: Trading
  - name: Orderbook
  - name: Milestones
  - name: Vault
paths:
  /orders:
    post:
      tags:
        - Orders
      summary: Create order
      description: >-
        Build an unsigned transaction that buys or sells YES or NO contracts on
        a market. Send exactly one quantity field: `contractsMicro` (1,000,000 =
        1 contract), `contractsDecimal`, or the legacy whole-number `contracts`.
        Sign the returned transaction and submit it to `POST /execute` together
        with the `execution.context` object, unchanged.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '200':
          description: Unsigned transaction for creating an order
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateOrderResponse'
        '400':
          description: Invalid request parameters
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    CreateOrderRequest:
      type: object
      properties:
        ownerPubkey:
          type: string
          description: Order owner public key
        marketId:
          type: string
          minLength: 1
          description: Hashed or canonical market ID for buys
        positionPubkey:
          type: string
          minLength: 32
          description: Required for sells
        isYes:
          type: boolean
          description: '`true` to trade YES contracts, `false` for NO.'
        isBuy:
          type: boolean
          description: '`true` to buy contracts, `false` to sell.'
        contracts:
          anyOf:
            - type: string
            - type: number
          description: >-
            Legacy whole-contract sell quantity (only when `isBuy` is `false`).
            Cannot express fractions; prefer contractsMicro or contractsDecimal.
            Provide only one of contracts, contractsMicro, or contractsDecimal.
        contractsMicro:
          anyOf:
            - type: string
            - type: number
          description: >-
            Exact sell quantity in micro-contract units (1000000 = 1 contract).
            Provide only one of contracts, contractsMicro, or contractsDecimal.
        contractsDecimal:
          anyOf:
            - type: string
            - type: number
          description: >-
            Exact sell quantity as a decimal (e.g. 5.777773). Provide only one
            of contracts, contractsMicro, or contractsDecimal.
        depositAmount:
          anyOf:
            - type: string
            - type: number
          description: >-
            Amount must conform to the token decimals used by the token mint in
            `depositMint`
        depositMint:
          type: string
          default: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
          description: Mint address for the deposit token
      required:
        - isBuy
    CreateOrderResponse:
      type: object
      properties:
        transaction:
          type: string
          nullable: true
          description: Base64 encoded transaction
        txMeta:
          type: object
          nullable: true
          properties:
            blockhash:
              type: string
            lastValidBlockHeight:
              type: integer
              minimum: 0
          required:
            - blockhash
            - lastValidBlockHeight
          description: Transaction metadata
        externalOrderId:
          type: string
          nullable: true
        requiredSigners:
          type: array
          items:
            type: string
          description: Public keys that must sign the returned transaction
        execution:
          $ref: '#/components/schemas/Execution'
        executionModel:
          type: string
          nullable: true
          description: >-
            Execution model for the order. `atomic_swap` for Jupiter Forecast
            (bisonfi) orders, which execute as a single atomic swap through
            `POST /execute`. `null` for keeper-filled prediction orders.
        settlement:
          type: string
          nullable: true
          description: >-
            Settlement mode. `auto` for Jupiter Forecast orders, which settle
            automatically. `null` for keeper-filled prediction orders.
        jupiterSwapRequestId:
          type: string
          nullable: true
          description: >-
            Jupiter swap request id when the order executes as a swap (Jupiter
            Forecast). `null` otherwise.
        order:
          type: object
          properties:
            orderPubkey:
              type: string
              nullable: true
              description: Order account public key
            orderAtaPubkey:
              type: string
              nullable: true
              description: Order escrow token account
            userPubkey:
              type: string
              description: Order owner public key
            marketId:
              type: string
              description: Associated market identifier
            marketIdHash:
              type: string
              description: Hashed market identifier used for on-chain PDAs
            positionPubkey:
              type: string
              description: Position PDA used for the order
            isBuy:
              type: boolean
              description: True when the order increases exposure
            isYes:
              type: boolean
              description: True when the order is on the YES side
            contracts:
              type: string
              description: >-
                Legacy whole-contract quantity. Use contractsMicro or
                contractsDecimal for the exact fractional quantity.
            contractsMicro:
              type: string
              description: >-
                Contracts in micro-contract units (u64 as string, 1000000 = 1
                contract)
            contractsDecimal:
              type: string
              description: Contracts as a decimal string
            newContracts:
              type: string
              description: >-
                Position contracts after executing the order (legacy
                whole-contract quantity)
            newContractsMicro:
              type: string
              description: Position contracts after the order, in micro-contract units
            newContractsDecimal:
              type: string
              description: Position contracts after the order, as a decimal string
            maxBuyPriceUsd:
              type: string
              nullable: true
              description: Max fill price for buys (micro USD)
            minSellPriceUsd:
              type: string
              nullable: true
              description: Min fill price for sells (micro USD)
            externalOrderId:
              type: string
              nullable: true
              description: Client provided order identifier
            orderCostUsd:
              type: string
              description: Notional for this order in micro USD
            newAvgPriceUsd:
              type: string
              description: Average position entry price after the order
            newSizeUsd:
              type: string
              description: Total position cost basis after the order
            newPayoutUsd:
              type: string
              description: Position max payout after the order
            estimatedProtocolFeeUsd:
              type: string
              description: Estimated protocol fee (micro USD)
            estimatedVenueFeeUsd:
              type: string
              description: Estimated venue fee (micro USD)
            estimatedTotalFeeUsd:
              type: string
              description: Estimated total fees applied (micro USD)
            payoutUsd:
              type: string
              description: Position max payout for this order (micro USD)
            slippageBps:
              type: integer
              minimum: 0
              description: >-
                Estimated slippage for this order, in basis points (0 = none,
                250 = 2.5%). For a Polymarket buy this is the estimated price
                impact from sweeping orderbook levels; for a sell it is the max
                allowed slippage from best bid to the protected sell floor.
                Treat it as a quote signal, not a guaranteed loss.
            maxSlippageBps:
              type: integer
              minimum: 0
              nullable: true
              description: Maximum slippage allowed for this order, in basis points
          required:
            - orderPubkey
            - orderAtaPubkey
            - userPubkey
            - marketId
            - marketIdHash
            - positionPubkey
            - isBuy
            - isYes
            - contracts
            - newContracts
            - maxBuyPriceUsd
            - minSellPriceUsd
            - externalOrderId
            - orderCostUsd
            - newAvgPriceUsd
            - newSizeUsd
            - newPayoutUsd
            - estimatedProtocolFeeUsd
            - estimatedVenueFeeUsd
            - estimatedTotalFeeUsd
      required:
        - transaction
        - txMeta
        - externalOrderId
        - order
    ErrorResponse:
      type: object
      properties:
        type:
          type: string
          enum:
            - invalid_request_error
            - authentication_error
            - permission_error
            - idempotency_error
            - rate_limit_error
            - api_error
        message:
          type: string
        code:
          type: string
        param:
          type: string
        request_id:
          type: string
        doc_url:
          type: string
      required:
        - type
        - message
        - request_id
    Execution:
      type: object
      description: >-
        Execution details returned with an order build. Sign the build's
        `transaction`, then pass `context` unchanged to `POST /execute`.
      properties:
        endpoint:
          type: string
          description: >-
            Relative execute path. Resolves to
            `https://api.jup.ag/prediction/v1/execute`.
          example: /api/v1/execute
        context:
          type: object
          additionalProperties: true
          description: >-
            Opaque execution context. Pass this object unchanged to `POST
            /execute`. Includes a `type` field (`create_order` for keeper-filled
            orders, `bisonfi_swap` for Jupiter Forecast orders).
      required:
        - endpoint
        - context
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.