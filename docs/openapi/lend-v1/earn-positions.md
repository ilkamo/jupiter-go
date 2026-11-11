> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Earn Positions

> Request for the position data of one or multiple users



## OpenAPI

````yaml openapi-spec/lend/lend.yaml get /earn/positions
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
  /earn/positions:
    get:
      summary: positions
      description: |
        Request for the position data of one or multiple users
      parameters:
        - name: users
          in: query
          description: |
            User wallet addresses (comma separated)
          required: true
          schema:
            type: string
          example: >-
            HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf,jdocuPgEAjMfihABsPgKEvYtsmMzjUHeq9LX4Hvs7f3
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserPositionsResponse'
              example:
                - token:
                    id: So11111111111111111111111111111111111111112
                    address: So11111111111111111111111111111111111111112
                    name: Wrapped SOL
                    symbol: SOL
                    decimals: 9
                  ownerAddress: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
                  shares: '100000000'
                  underlyingAssets: '103448300'
                  underlyingBalance: '17.72'
                  allowance: '1000000000'
      security:
        - ApiKeyAuth: []
components:
  schemas:
    UserPositionsResponse:
      type: array
      items:
        $ref: '#/components/schemas/UserPosition'
    UserPosition:
      type: object
      required:
        - token
        - ownerAddress
        - shares
        - underlyingAssets
        - underlyingBalance
        - allowance
      properties:
        token:
          $ref: '#/components/schemas/TokenInfo'
        ownerAddress:
          type: string
          description: |
            - Address of user's wallet
        shares:
          type: string
          description: |
            - Balance of user's jlTokens
        underlyingAssets:
          type: string
          description: |
            - Balance of user's USDC in protocol (including intereste accured)
        underlyingBalance:
          type: string
          description: |
            - Balance of user's USDC in wallet
        allowance:
          type: string
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