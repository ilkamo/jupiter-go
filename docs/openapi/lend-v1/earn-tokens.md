> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Tokens

> Request for the tokens available to be deposited and their information



## OpenAPI

````yaml openapi-spec/lend/lend.yaml get /earn/tokens
openapi: 3.0.3
info:
  title: Jupiter Lend API
  version: 1.0.0
  description: Jupiter Lend API Schema
servers:
  - url: https://api.jup.ag/lend/v1
    description: Jupiter Lend API endpoint
security: []
paths:
  /earn/tokens:
    get:
      summary: tokens
      description: |
        Request for the tokens available to be deposited and their information
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokensResponse'
              example:
                - id: So11111111111111111111111111111111111111112
                  address: So11111111111111111111111111111111111111112
                  name: Wrapped SOL
                  symbol: SOL
                  decimals: 9
                  assetAddress: So11111111111111111111111111111111111111112
                  asset:
                    address: So11111111111111111111111111111111111111112
                    chain_id: 0
                    name: Wrapped SOL
                    symbol: SOL
                    decimals: 9
                    logo_url: >-
                      https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/So11111111111111111111111111111111111111112/logo.png
                    price: 171.34
                    coingecko_id: solana
                  totalAssets: '1500000000000'
                  totalSupply: '1450000000000'
                  convertToShares: '0.966667'
                  convertToAssets: '1.034483'
                  rewardsRate: '0.035'
                  supplyRate: '0.025'
                  totalRate: '0.06'
      security:
        - ApiKeyAuth: []
components:
  schemas:
    TokensResponse:
      type: array
      items:
        $ref: '#/components/schemas/TokenInfo'
    TokenInfo:
      type: object
      required:
        - id
        - address
        - name
        - symbol
        - decimals
        - assetAddress
        - asset
        - totalAssets
        - totalSupply
        - convertToShares
        - convertToAssets
        - rewardsRate
        - supplyRate
        - totalRate
        - rebalanceDifference
        - liquiditySupplyData
      properties:
        id:
          type: integer
          description: |
            - Lending ID, starts from `1`
        address:
          type: string
          description: >
            - Lending token address like
            `9BEcn9aPEmhSPbPQeFGjidRiEKki46fVQDyPpSQXPA2D`
        name:
          type: string
          description: |
            - Lending token name like Jupiter Lend USDC
        symbol:
          type: string
          description: |
            - Lending token symbol like jlUSDC
        decimals:
          type: integer
        assetAddress:
          type: string
          description: >
            - Lending token underlying asset address like
            `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`
        asset:
          $ref: '#/components/schemas/AssetInfo'
        totalAssets:
          type: string
          description: >
            - Total available underlying mint from jlTokens (with accured
            interest)
        totalSupply:
          type: string
          description: |
            - Total supply of jlTokens for this mint
        convertToShares:
          type: string
          description: |
            - Value of one 1 underlying token to jlToken, like 1 USDC = X jlUSDC
            - Used for conversion
        convertToAssets:
          type: string
          description: >
            - Value of one 1 jlToken token to underlying mint, like 1 jlUSDC = X
            USDC

            - Used for conversion
        rewardsRate:
          type: string
          description: |
            - Rewards rate , in 1e4 decimals
        supplyRate:
          type: string
          description: |
            - Supply rate, in 1e4 decimals
        totalRate:
          type: string
          description: |
            - Total sum of rewards + supply rate
        rebalanceDifference:
          type: string
        liquiditySupplyData:
          $ref: '#/components/schemas/LiquiditySupplyData'
    AssetInfo:
      type: object
      required:
        - address
        - chain_id
        - name
        - symbol
        - decimals
        - logo_url
        - price
        - coingecko_id
      properties:
        address:
          type: string
          description: >
            - Lending token underlying asset address like
            `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`
        chain_id:
          type: string
          description: |
            - Lending token underlying asset chain id like `solana`
        name:
          type: string
        symbol:
          type: string
        decimals:
          type: integer
        logo_url:
          type: string
          format: uri
        price:
          type: string
        coingecko_id:
          type: string
    LiquiditySupplyData:
      type: object
      required:
        - modeWithInterest
        - supply
        - withdrawalLimit
        - lastUpdateTimestamp
        - expandPercent
        - expandDuration
        - baseWithdrawalLimit
        - withdrawableUntilLimit
        - withdrawable
      properties:
        modeWithInterest:
          type: boolean
          description: |
            - The underlying asset is earning interest at liquidity layer
        supply:
          type: string
          description: |
            - Total supply of underlying asset on liquidity layer
            - In decimals, like 3168044480191 / 1e6 for USDC
        withdrawalLimit:
          type: string
          description: |
            - Available withdrawal limit for this asset
        lastUpdateTimestamp:
          type: string
          description: |
            - When was the reserve last updated
        expandPercent:
          type: string
        expandDuration:
          type: string
        baseWithdrawalLimit:
          type: string
        withdrawableUntilLimit:
          type: string
        withdrawable:
          type: string
          description: >
            - Total amount withdrawable at this current point, it expands every
            second.
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.