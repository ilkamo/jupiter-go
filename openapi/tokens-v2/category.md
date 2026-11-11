> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Category

> Returns an array of mints and their information for the given category and interval



## OpenAPI

````yaml openapi-spec/tokens/v2/tokens.yaml get /{category}/{interval}
openapi: 3.0.3
info:
  title: Jupiter Tokens API V2
  version: 2.0.0
  description: Jupiter Tokens API V2 Schema
servers:
  - url: https://api.jup.ag/tokens/v2
    description: Jupiter Tokens API V2 Endpoint
security: []
paths:
  /{category}/{interval}:
    get:
      summary: category
      description: >
        Returns an array of mints and their information for the given category
        and interval
      parameters:
        - in: path
          name: category
          required: true
          schema:
            type: string
            enum:
              - toporganicscore
              - toptraded
              - toptrending
          description: |
            - Top tokens in different trading categories
            - The result filters out generic top tokens like SOL, USDC, etc
            - Default to 50 mints in response
        - in: path
          name: interval
          required: true
          schema:
            type: string
            enum:
              - 5m
              - 1h
              - 6h
              - 24h
          description: |
            - Query by time interval for more accuracy
        - in: query
          name: limit
          schema:
            type: integer
            minimum: 1
            maximum: 100
          required: false
          description: |
            - Maximum number of results to return (default 50, max 100)
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/MintInformation'
              example:
                - id: So11111111111111111111111111111111111111112
                  name: Wrapped SOL
                  symbol: SOL
                  decimals: 9
                  usdPrice: 171.34
                  stats24h:
                    priceChange: 2.5
                    volumeChange: 15000000
                    numBuys: 50000
                    numSells: 48000
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
      security:
        - ApiKeyAuth: []
components:
  schemas:
    MintInformation:
      type: object
      properties:
        id:
          type: string
          description: The token's mint address
        name:
          type: string
        symbol:
          type: string
        icon:
          type: string
          nullable: true
          description: URL to the token's logo image
        decimals:
          type: number
        tokenProgram:
          type: string
          description: The token program address (SPL Token or Token-2022)
        createdAt:
          type: string
          format: date-time
          description: Token creation timestamp
        twitter:
          type: string
          nullable: true
        telegram:
          type: string
          nullable: true
        website:
          type: string
          nullable: true
        discord:
          type: string
          nullable: true
        instagram:
          type: string
          nullable: true
        tiktok:
          type: string
          nullable: true
        otherUrl:
          type: string
          nullable: true
        dev:
          type: string
          nullable: true
          description: The token's developer wallet address
        mintAuthority:
          type: string
          nullable: true
          description: Mint authority address (present when mint authority is not disabled)
        freezeAuthority:
          type: string
          nullable: true
          description: >-
            Freeze authority address (present when freeze authority is not
            disabled)
        circSupply:
          type: number
          nullable: true
        totalSupply:
          type: number
          nullable: true
        launchpad:
          type: string
          nullable: true
        partnerConfig:
          type: string
          nullable: true
        graduatedPool:
          type: string
          nullable: true
        graduatedAt:
          type: string
          nullable: true
        holderCount:
          type: number
          nullable: true
        fdv:
          type: number
          nullable: true
          description: Fully diluted valuation in USD
        mcap:
          type: number
          nullable: true
          description: Market cap in USD
        usdPrice:
          type: number
          nullable: true
        priceBlockId:
          type: number
          nullable: true
          description: Solana block number for the price data point
        liquidity:
          type: number
          nullable: true
          description: Total liquidity in USD
        apy:
          type: object
          nullable: true
          description: APY data, only present for assets listed on Jupiter Lend's Earn
          properties:
            jupEarn:
              type: number
              description: Annual percentage yield from Jupiter Earn
        stats5m:
          $ref: '#/components/schemas/SwapStats'
          nullable: true
        stats1h:
          $ref: '#/components/schemas/SwapStats'
          nullable: true
        stats6h:
          $ref: '#/components/schemas/SwapStats'
          nullable: true
        stats24h:
          $ref: '#/components/schemas/SwapStats'
          nullable: true
        firstPool:
          type: object
          nullable: true
          properties:
            id:
              type: string
              description: Pool address
            createdAt:
              type: string
              description: Pool creation timestamp
        audit:
          type: object
          nullable: true
          description: >
            Audit information. All sub-fields are conditional and vary per
            token.
          properties:
            isSus:
              type: boolean
              nullable: true
              description: >
                Flagged as suspicious. This field is only present when the token

                has been flagged — check for the field's presence, not its
                value.

                Not all risky tokens will have this flag.
            mintAuthorityDisabled:
              type: boolean
              nullable: true
            freezeAuthorityDisabled:
              type: boolean
              nullable: true
            topHoldersPercentage:
              type: number
              nullable: true
              description: |
                Percentage of supply held by the top holders, on a 0-100 scale
                (e.g. 25.07 means 25.07%, not the fraction 0.2507).
            devBalancePercentage:
              type: number
              nullable: true
              description: |
                Percentage of supply held by the developer, on a 0-100 scale
                (e.g. 2.5 means 2.5%, not the fraction 0.025).
            devMints:
              type: number
              nullable: true
              description: Number of developer mint events
        organicScore:
          type: number
          description: Organic trading activity score (0-100)
        organicScoreLabel:
          type: string
          enum:
            - high
            - medium
            - low
        isVerified:
          type: boolean
          nullable: true
        tags:
          type: array
          items:
            type: string
          nullable: true
        updatedAt:
          type: string
          format: date-time
          description: Last data update timestamp
    SwapStats:
      type: object
      properties:
        priceChange:
          type: number
          nullable: true
        holderChange:
          type: number
          nullable: true
        liquidityChange:
          type: number
          nullable: true
        volumeChange:
          type: number
          nullable: true
        buyVolume:
          type: number
          nullable: true
        sellVolume:
          type: number
          nullable: true
        buyOrganicVolume:
          type: number
          nullable: true
        sellOrganicVolume:
          type: number
          nullable: true
        numBuys:
          type: number
          nullable: true
        numSells:
          type: number
          nullable: true
        numTraders:
          type: number
          nullable: true
        numOrganicBuyers:
          type: number
          nullable: true
        numNetBuyers:
          type: number
          nullable: true
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.