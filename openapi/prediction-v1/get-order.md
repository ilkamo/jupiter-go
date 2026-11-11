> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Order (Prediction API)

> Get detailed information about a specific order



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /orders/{orderPubkey}
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
  /orders/{orderPubkey}:
    get:
      tags:
        - Orders
      summary: Get order
      description: >-
        One order by its account public key: market, side, size, status and
        timestamps.
      parameters:
        - schema:
            type: string
            minLength: 1
            description: Order account public key
          required: true
          description: Order account public key
          name: orderPubkey
          in: path
      responses:
        '200':
          description: Order data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Order'
        '400':
          description: Failed to fetch order
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Order not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    Order:
      type: object
      properties:
        pubkey:
          type: string
          description: Order account public key
        owner:
          type: string
          description: Order owner public key
        ownerPubkey:
          type: string
          description: >-
            Order owner public key (alias of owner, use ownerPubkey when
            available)
        market:
          type: string
          description: Associated market public key
        marketId:
          type: string
          description: External market identifier used to derive the PDA
        marketIdHash:
          type: string
          description: Hashed market identifier used for on-chain PDAs
        eventId:
          type: string
          description: External event identifier
        position:
          type: string
          description: Associated position public key
        status:
          type: string
          enum:
            - pending
            - filled
            - failed
          description: Current order status
        isYes:
          type: boolean
          description: True when order is for the YES side
        isBuy:
          type: boolean
          description: True when order is a buy order
        createdAt:
          type: integer
          description: Unix timestamp (seconds) when the order was created
        updatedAt:
          type: integer
          description: Unix timestamp (seconds) when the order last changed on-chain
        contracts:
          type: string
          description: Number of contracts (u64 as string)
        maxFillPriceUsd:
          type: string
          description: Maximum fill price in micro USD (u64 as string)
        maxBuyPriceUsd:
          type: string
          nullable: true
          description: Buyer-specified max fill price (micro USD)
        minSellPriceUsd:
          type: string
          nullable: true
          description: Seller-specified min fill price (micro USD)
        filledAt:
          type: integer
          description: Unix timestamp (seconds) when the order was filled (0 when pending)
        filledContracts:
          type: string
          description: Number of filled contracts (u64 as string)
        avgFillPriceUsd:
          type: string
          description: Average fill price in micro USD (u64 as string)
        settled:
          type: boolean
          description: Whether the order has been settled on-chain
        orderId:
          type: string
          description: External order identifier from the venue
        sizeUsd:
          type: string
          description: Order notional in micro USD (u128 as string)
        eventMetadata:
          $ref: '#/components/schemas/EventMetadata'
        marketMetadata:
          allOf:
            - $ref: '#/components/schemas/MarketMetadata'
            - properties:
                marketId:
                  type: string
                eventId:
                  type: string
                title:
                  type: string
                status:
                  type: string
                result:
                  type: string
                closeTime:
                  type: number
                openTime:
                  type: number
                isTeamMarket:
                  type: boolean
        externalOrderId:
          type: string
          description: Client-provided order identifier
        bump:
          type: integer
          minimum: 0
      required:
        - pubkey
        - owner
        - ownerPubkey
        - market
        - marketId
        - marketIdHash
        - eventId
        - position
        - status
        - isYes
        - isBuy
        - createdAt
        - updatedAt
        - contracts
        - maxFillPriceUsd
        - maxBuyPriceUsd
        - minSellPriceUsd
        - filledAt
        - filledContracts
        - avgFillPriceUsd
        - settled
        - orderId
        - sizeUsd
        - eventMetadata
        - marketMetadata
        - externalOrderId
        - bump
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
    EventMetadata:
      type: object
      properties:
        eventId:
          type: string
        title:
          type: string
        subtitle:
          type: string
        slug:
          type: string
        series:
          type: string
        closeTime:
          type: string
          description: ISO 8601 formatted close time
        imageUrl:
          type: string
          format: uri
        isLive:
          type: boolean
      required:
        - eventId
    MarketMetadata:
      type: object
      properties:
        marketId:
          type: string
        eventId:
          type: string
        title:
          type: string
        subtitle:
          type: string
        description:
          type: string
        provider:
          type: string
        status:
          type: string
        result:
          type: string
          nullable: true
        closeTime:
          type: number
        openTime:
          type: number
        isTeamMarket:
          type: boolean
        marketOptions:
          type: array
          items:
            type: object
            properties:
              label:
                type: string
              buyYes:
                type: boolean
        rulesPrimary:
          type: string
        rulesSecondary:
          type: string
      required:
        - marketId
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.