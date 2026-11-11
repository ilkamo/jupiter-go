> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Search

> Request a search by token's symbol, name or mint address



## OpenAPI

````yaml openapi-spec/tokens/v2/tokens.yaml get /search
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
  /search:
    get:
      summary: search
      description: |
        Request a search by token's symbol, name or mint address
      parameters:
        - in: query
          name: query
          schema:
            type: string
          required: true
          description: >
            - Search for a token and its information by its symbol, name or mint
            address

            - Comma-separate to ONLY search for multiple mint addresses

            - Limit to 100 mint addresses in query

            - Default to 20 mints in response when searching via symbol or name
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
                - id: JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN
                  name: Jupiter
                  symbol: JUP
                  icon: https://static.jup.ag/jup/icon.png
                  decimals: 6
                  tokenProgram: TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA
                  createdAt: '2024-06-07T10:56:42.584Z'
                  twitter: https://twitter.com/JupiterExchange
                  website: https://jup.ag
                  dev: JUPhop9E8ZfdJ5FNHhxQt4uAih822Vs4QpqsWcewFbq
                  circSupply: 3243891294.88
                  totalSupply: 6863982654.38
                  firstPool:
                    id: 2pspvjWWaf3dNgt3jsgSzFCNvMGPb7t8FrEYvLGjvcCe
                    createdAt: '2024-01-29T17:33:29Z'
                  holderCount: 857254
                  audit:
                    mintAuthorityDisabled: true
                    freezeAuthorityDisabled: true
                    topHoldersPercentage: 15.45
                    devMints: 1
                  organicScore: 97.98
                  organicScoreLabel: high
                  isVerified: true
                  tags:
                    - community
                    - strict
                    - verified
                  fdv: 971593450.82
                  mcap: 459171256.1
                  usdPrice: 0.1415
                  priceBlockId: 402118601
                  liquidity: 2992387.89
                  stats24h:
                    priceChange: -8.27
                    holderChange: -0.03
                    liquidityChange: -6.96
                    volumeChange: 30.39
                    buyVolume: 1821969.64
                    sellVolume: 2016195.15
                    numBuys: 33979
                    numSells: 31608
                    numTraders: 5174
                  updatedAt: '2026-02-23T04:55:12.461Z'
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