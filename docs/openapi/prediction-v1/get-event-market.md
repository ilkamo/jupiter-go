> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Event Market

> Get detailed market information for a specific market within an event



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /events/{eventId}/markets/{marketId}
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
  /events/{eventId}/markets/{marketId}:
    get:
      tags:
        - Markets
      summary: Get event market
      description: >-
        One market scoped to its event: title, status, result, open and close
        times, resolution time, rules and outcomes with current pricing.
      parameters:
        - schema:
            type: string
            minLength: 1
            description: Event identifier
          required: true
          description: Event identifier
          name: eventId
          in: path
        - schema:
            type: string
            minLength: 1
            description: Market identifier
          required: true
          description: Market identifier
          name: marketId
          in: path
      responses:
        '200':
          description: Market data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Market'
        '404':
          description: Market not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    Market:
      type: object
      description: Market fields are returned flat (there is no nested `metadata` object).
      properties:
        marketId:
          type: string
          description: Market identifier
        eventId:
          type: string
          description: Parent event identifier
        provider:
          type: string
          enum:
            - polymarket
            - gx
            - bisonfi
          description: Market data provider
        title:
          type: string
          description: Market title
        status:
          type: string
          enum:
            - open
            - closed
            - cancelled
          description: Current market status
        result:
          type: string
          nullable: true
          enum:
            - 'yes'
            - 'no'
          description: Market result. null until resolved, then "yes" or "no"
        openTime:
          type: number
          description: Unix timestamp (seconds) when the market opens
        closeTime:
          type: number
          description: Unix timestamp (seconds) when the market closes
        resolveAt:
          nullable: true
          oneOf:
            - type: string
            - type: number
          description: >-
            When the market resolved. null while open; returned as an ISO 8601
            string once resolved
        marketResultPubkey:
          type: string
          nullable: true
          description: Public key of the market result account
        imageUrl:
          type: string
          nullable: true
          format: uri
          description: Market image URL
        rulesPrimary:
          type: string
          description: Primary resolution rules
        rulesSecondary:
          type: string
          description: Secondary resolution rules
        outcomes:
          type: array
          items:
            type: string
          description: Outcome labels for the market
        marketOptions:
          type: array
          items:
            type: object
            properties:
              label:
                type: string
              buyYes:
                type: boolean
          description: Selectable options for team or multi-outcome markets
        clobTokenIds:
          type: array
          items:
            type: string
          description: Underlying CLOB token IDs (provider-specific)
        isTeamMarket:
          type: boolean
          description: True for team or sports markets
        team:
          type: object
          nullable: true
          description: Team metadata, present for team markets
        sportsLine:
          type: number
          nullable: true
          description: Sports line or handicap, when applicable
        sportsMarketType:
          type: string
          nullable: true
          description: Sports market type, when applicable
        outcomeMint:
          type: string
          nullable: true
          description: >-
            Outcome token mint for this market side (Jupiter Forecast /
            bisonfi). Buying the market gives you this token; it can also be
            traded directly through the Swap API.
        outcomeTokenProgram:
          type: string
          nullable: true
          description: >-
            Token program that owns `outcomeMint`. Jupiter Forecast uses
            Token-2022 (`TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb`).
        marketPda:
          type: string
          nullable: true
          description: On-chain market account (Jupiter Forecast / bisonfi).
        outcomeSide:
          type: string
          nullable: true
          description: Outcome side for Jupiter Forecast markets (`up` or `down`).
        tradable:
          type: boolean
          description: >-
            Whether this Jupiter Forecast side can be traded right now. The
            Discover example filters on this field.
        lifecycleStatus:
          type: string
          enum:
            - open
            - resolving
            - settled
          description: >-
            Jupiter Forecast trading lifecycle: `open`, `resolving` after the
            close time, or `settled` after issuer resolution. Only `open` is
            tradable.
        pricing:
          $ref: '#/components/schemas/MarketPricing'
      required:
        - marketId
        - status
        - result
        - openTime
        - closeTime
        - resolveAt
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
    MarketPricing:
      type: object
      properties:
        buyYesPriceUsd:
          type: number
          nullable: true
        buyNoPriceUsd:
          type: number
          nullable: true
        sellYesPriceUsd:
          type: number
          nullable: true
        sellNoPriceUsd:
          type: number
          nullable: true
        volume:
          type: number
          description: Trading volume for this market
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.