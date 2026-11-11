> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Positions (Prediction API)

> Get a list of positions with optional filtering



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /positions
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
  /positions:
    get:
      tags:
        - Positions
      summary: List positions
      description: >-
        Paginated list of onchain positions, each holding YES or NO contracts in
        one market. Filter by `ownerPubkey`, `marketPubkey`, `marketId` and
        `isYes`; paginate with `start` and `end`. USD values are in micro USD
        (1,000,000 = $1).
      parameters:
        - schema:
            type: integer
            nullable: true
            minimum: 0
          required: false
          description: Pagination start index of the page to return.
          name: start
          in: query
        - schema:
            type: integer
            nullable: true
            minimum: 0
          required: false
          description: Pagination end index of the page to return.
          name: end
          in: query
        - schema:
            type: string
            minLength: 1
            description: Position owner public key
          required: false
          description: Position owner public key
          name: ownerPubkey
          in: query
        - schema:
            type: string
            minLength: 1
          required: false
          description: Filter by market account public key.
          name: marketPubkey
          in: query
        - schema:
            type: string
            minLength: 1
          required: false
          description: Filter by market id.
          name: marketId
          in: query
        - schema:
            type: string
            enum:
              - 'true'
              - 'false'
            description: Filter by position side (true = YES, false = NO)
          required: false
          description: Filter by position side (true = YES, false = NO)
          name: isYes
          in: query
      responses:
        '200':
          description: List of on-chain positions
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Position'
                  pagination:
                    type: object
                    properties:
                      start:
                        type: integer
                        minimum: 0
                      end:
                        type: integer
                        minimum: 0
                      total:
                        type: integer
                        minimum: 0
                      hasNext:
                        type: boolean
                    required:
                      - start
                      - end
                      - total
                      - hasNext
                required:
                  - data
                  - pagination
        '400':
          description: Failed to fetch positions
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    Position:
      type: object
      properties:
        pubkey:
          type: string
          description: Position account public key
        owner:
          type: string
          description: Position owner public key
        ownerPubkey:
          type: string
          description: >-
            Position owner public key (alias of owner, prefer ownerPubkey going
            forward)
        market:
          type: string
          description: Deterministic market PDA derived from marketId
        marketId:
          type: string
          description: External market identifier (ticker)
        marketIdHash:
          type: string
          description: Hashed market identifier used for on-chain PDAs
        isYes:
          type: boolean
          description: True if this is a YES position
        contracts:
          type: string
          description: >-
            Legacy whole-contract quantity held. Use contractsMicro or
            contractsDecimal for the exact fractional quantity.
        contractsMicro:
          type: string
          description: >-
            Contracts held in micro-contract units (u64 as string, 1000000 = 1
            contract)
        contractsDecimal:
          type: string
          description: Contracts held as a decimal string
        source:
          type: string
          description: Origin of the position (e.g. "program")
        maxSlippageBps:
          type: integer
          minimum: 0
          description: Maximum slippage recorded for the position, in basis points
        totalCostUsd:
          type: string
          description: Total cost basis in micro USD (u128 as string)
        sizeUsd:
          type: string
          description: Alias of total cost basis in micro USD (u128 as string)
        valueUsd:
          type: string
          nullable: true
          description: >-
            Mark-to-market value in micro USD (u128 as string); null once market
            is closed
        avgPriceUsd:
          type: string
          description: Average entry price per contract in micro USD (u64 as string)
        markPriceUsd:
          type: string
          nullable: true
          description: >-
            Current mark price per contract in micro USD (u64 as string); null
            once market is closed
        sellPriceUsd:
          type: string
          nullable: true
          description: >-
            Current best exit price for this side (micro USD string), or null
            when unavailable
        pnlUsd:
          type: string
          nullable: true
          description: >-
            Unrealized PnL in micro USD (i128 as string); null once market is
            closed
        pnlUsdPercent:
          type: number
          nullable: true
          description: >-
            Unrealized PnL percentage relative to size; null once market is
            closed
        pnlUsdAfterFees:
          type: string
          nullable: true
          description: >-
            Unrealized PnL after fees in micro USD (i128 as string); null once
            market is closed
        pnlUsdAfterFeesPercent:
          type: number
          nullable: true
          description: >-
            Unrealized PnL after fees percentage relative to size; null once
            market is closed
        openOrders:
          type: integer
          minimum: 0
          description: Number of open orders
        feesPaidUsd:
          type: string
          description: Total fees paid in micro USD (u64 as string)
        realizedPnlUsd:
          type: number
          description: Realized PnL in micro USD (i64 as number)
        claimed:
          type: boolean
          description: Whether payout has been claimed
        claimedUsd:
          type: string
          description: Amount claimed in micro USD (u64 as string)
        openedAt:
          type: integer
          description: Unix timestamp when the position was opened
        updatedAt:
          type: integer
          description: Unix timestamp of the last position update
        claimableAt:
          type: integer
          nullable: true
          description: >-
            Unix timestamp (seconds) when the position becomes claimable; null
            when unavailable
        payoutUsd:
          type: string
          description: Notional payout if position wins (contracts * $1)
        bump:
          type: integer
          minimum: 0
        eventId:
          type: string
          description: External event identifier if available
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
        settlementDate:
          type: integer
          nullable: true
          description: >-
            Unix timestamp (seconds) when the market settles; null when
            unavailable
        claimable:
          type: boolean
          description: True when the position qualifies for payout claim
      required:
        - pubkey
        - owner
        - ownerPubkey
        - market
        - marketId
        - marketIdHash
        - isYes
        - contracts
        - totalCostUsd
        - sizeUsd
        - valueUsd
        - avgPriceUsd
        - markPriceUsd
        - sellPriceUsd
        - pnlUsd
        - pnlUsdPercent
        - pnlUsdAfterFees
        - pnlUsdAfterFeesPercent
        - openOrders
        - feesPaidUsd
        - realizedPnlUsd
        - claimed
        - claimedUsd
        - openedAt
        - updatedAt
        - claimableAt
        - payoutUsd
        - bump
        - eventId
        - eventMetadata
        - marketMetadata
        - settlementDate
        - claimable
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