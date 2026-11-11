> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get History

> Get trading history and event records for an account



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /history
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
  /history:
    get:
      tags:
        - History
      summary: Get history
      description: >-
        Paginated history of order, fill, settlement and claim events. Filter by
        `ownerPubkey` or `positionPubkey`, or pass `id` to fetch a single event.
        Paginate with `start` and `end`.
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
            minLength: 32
            description: History owner public key
          required: false
          description: History owner public key
          name: ownerPubkey
          in: query
        - schema:
            type: integer
            minimum: 1
            description: History event ID
          required: false
          description: History event ID
          name: id
          in: query
        - schema:
            type: string
            minLength: 32
            description: Filter by position public key
          required: false
          description: Filter by position public key
          name: positionPubkey
          in: query
      responses:
        '200':
          description: Paginated event history, or single event when id is provided
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/HistoryEvent'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
                required:
                  - data
                  - pagination
components:
  schemas:
    HistoryEvent:
      type: object
      properties:
        id:
          type: integer
          description: Unique history event identifier
        eventType:
          type: string
          enum:
            - order_created
            - order_filled
            - order_failed
            - order_closed
            - payout_claimed
            - position_updated
            - position_lost
          description: >-
            Type of history event. Position wins surface as payout_claimed
            (there is no position_won).
        signature:
          type: string
          description: Solana transaction signature
        slot:
          type: string
          description: Solana slot number (u64 as string)
        timestamp:
          type: integer
          description: Unix timestamp (seconds) when the event occurred
        orderPubkey:
          type: string
          description: Associated order account public key
        positionPubkey:
          type: string
          description: Associated position account public key
        marketId:
          type: string
          description: External market identifier
        ownerPubkey:
          type: string
          description: Position or order owner public key
        keeperPubkey:
          type: string
          description: Keeper that processed the transaction
        externalOrderId:
          type: string
          description: Client-provided order identifier
        orderId:
          type: string
          description: External order identifier from the venue
        isBuy:
          type: boolean
          description: True when the event relates to a buy order
        isYes:
          type: boolean
          description: True when the event relates to the YES side
        contracts:
          type: string
          description: >-
            Number of contracts in the order. Legacy whole-contract value; see
            contractsMicro/contractsDecimal for the exact quantity.
        contractsMicro:
          type: string
          description: >-
            Contracts in micro-contract units (u64 as string, 1000000 = 1
            contract)
        contractsDecimal:
          type: string
          description: Contracts as a decimal string
        filledContracts:
          type: string
          description: Number of contracts filled (legacy whole-contract value)
        filledContractsMicro:
          type: string
          description: Filled contracts in micro-contract units
        filledContractsDecimal:
          type: string
          description: Filled contracts as a decimal string
        contractsSettled:
          type: string
          description: >-
            Number of contracts settled in this event (legacy whole-contract
            value)
        contractsSettledMicro:
          type: string
          description: Settled contracts in micro-contract units
        contractsSettledDecimal:
          type: string
          description: Settled contracts as a decimal string
        maxFillPriceUsd:
          type: string
          description: Maximum fill price in micro USD (u64 as string)
        avgFillPriceUsd:
          type: string
          description: Average fill price in micro USD (u64 as string)
        maxBuyPriceUsd:
          type: string
          nullable: true
          description: Buyer-specified max fill price in micro USD (u64 as string)
        minSellPriceUsd:
          type: string
          nullable: true
          description: Seller-specified min fill price in micro USD (u64 as string)
        depositAmountUsd:
          type: string
          description: Amount deposited for this order in micro USD (u64 as string)
        totalCostUsd:
          type: string
          description: Total cost of the order in micro USD (u64 as string)
        feeUsd:
          type: string
          nullable: true
          description: Fee charged for this event in micro USD (u64 as string)
        grossProceedsUsd:
          type: string
          description: Gross proceeds before fees in micro USD (u64 as string)
        netProceedsUsd:
          type: string
          description: Net proceeds after fees in micro USD (u64 as string)
        transferAmountToken:
          type: string
          nullable: true
          description: >-
            Token amount transferred in this event (native token units as
            string)
        realizedPnl:
          type: string
          nullable: true
          description: Realized PnL from this event in micro USD (i64 as string)
        realizedPnlBeforeFees:
          type: string
          nullable: true
          description: Realized PnL before fees in micro USD (i64 as string)
        payoutAmountUsd:
          type: string
          description: >-
            Payout amount in micro USD (u64 as string), populated for
            payout_claimed events
        eventId:
          type: string
          description: External event identifier
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
          description: Metadata for the associated market
        eventMetadata:
          $ref: '#/components/schemas/EventMetadata'
          description: Metadata for the parent event
      required:
        - id
        - eventType
        - signature
        - slot
        - timestamp
        - orderPubkey
        - positionPubkey
        - marketId
        - ownerPubkey
        - keeperPubkey
        - externalOrderId
        - orderId
        - isBuy
        - isYes
        - contracts
        - filledContracts
        - contractsSettled
        - maxFillPriceUsd
        - avgFillPriceUsd
        - maxBuyPriceUsd
        - minSellPriceUsd
        - depositAmountUsd
        - totalCostUsd
        - feeUsd
        - grossProceedsUsd
        - netProceedsUsd
        - transferAmountToken
        - realizedPnl
        - realizedPnlBeforeFees
        - payoutAmountUsd
        - eventId
        - marketMetadata
        - eventMetadata
    Pagination:
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
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.