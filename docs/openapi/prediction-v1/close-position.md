> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Close Position

> Request an unsigned transaction to sell all contracts in a position



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml delete /positions/{positionPubkey}
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
  /positions/{positionPubkey}:
    delete:
      tags:
        - Positions
      summary: Close position
      description: >-
        Build an unsigned transaction that sells every contract of one position
        before the market resolves. Sign it and submit it to `POST /execute`
        together with the `execution.context` object.
      parameters:
        - schema:
            type: string
            minLength: 32
            description: Position account public key
          required: true
          description: Position account public key
          name: positionPubkey
          in: path
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ClosePositionRequest'
      responses:
        '200':
          description: Unsigned transaction that sells all contracts for a position
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateOrderResponse'
        '400':
          description: Unable to build close position transaction
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Position not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    ClosePositionRequest:
      type: object
      properties:
        ownerPubkey:
          type: string
          minLength: 32
          description: Position owner public key
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