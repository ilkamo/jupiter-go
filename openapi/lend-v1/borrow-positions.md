> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Borrow Positions

> Request the borrow positions of one or multiple users



## OpenAPI

````yaml openapi-spec/lend/borrow.yaml get /borrow/positions
openapi: 3.0.3
info:
  title: Jupiter Lend Borrow API
  description: >
    OpenAPI specification for Jupiter Lend borrow integrator endpoints.


    **Base URL:** `https://api.jup.ag/lend/v1`


    Share this file with documentation tools (ReadMe, Stoplight, Postman, Redoc,
    Mintlify,

    Swagger UI, etc.) to generate API reference docs.


    | Method | Path | Description |

    | ------ | ---- | ----------- |

    | GET | `/borrow/vaults` | List borrow vaults |

    | GET | `/borrow/positions` | List positions for wallet addresses |

    | POST | `/borrow/operate` | Build unsigned operate transaction |

    | POST | `/borrow/operate-instructions` | Build operate Solana instructions
    |


    **Notes**

    - Pass multiple wallets to `GET /borrow/positions` as a comma-separated
    `users` query value.

    - `colAmount` / `debtAmount` are signed integer strings in token base units.

    - Use `positionId: 0` in operate requests to open a new position.

    - Optional `market` query/body context: `main` (default) or `ethena`.
  version: 1.0.0
servers:
  - url: https://api.jup.ag/lend/v1
    description: Jupiter Lend API endpoint
security: []
tags:
  - name: Borrow
    description: Vault discovery, position lookup, and operate transaction building
paths:
  /borrow/positions:
    get:
      tags:
        - Borrow
      summary: List borrow positions for users
      description: >
        Returns borrow positions (NFTs) for one or more wallet addresses.


        Pass multiple wallets as a comma-separated list in the `users` query
        parameter.
      operationId: listBorrowPositions
      parameters:
        - name: users
          in: query
          description: Comma-separated Solana wallet addresses.
          required: true
          style: form
          explode: true
          schema:
            type: string
          example: >-
            HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf,TEsTKPSD2RBQxFHgxqmr3YLayAWWz2kpESk5wdoeB62
        - $ref: '#/components/parameters/market'
      responses:
        '200':
          description: Borrow positions for the requested users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/BorrowPosition'
              example:
                - id: 3518
                  vaultId: 8
                  address: 9tjzV49wccwf9A8J5fZ5josjWYnS41fUpNmcwqcqYrPF
                  supply: '0'
                  beforeSupply: '0'
                  borrow: '0'
                  beforeBorrow: '0'
                  isLiquidated: false
                  supplyLiquidation: '0'
                  borrowLiquidation: '0'
                  isSupplyPosition: true
                  tick: -2147483648
                  tickId: 1
                  dustBorrow: '0'
                  ownerAddress: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
                  vault:
                    id: 8
                    address: DchXYJwVjPaEUxCiySXVbNyNppAFbQM7Ku76uKvyaDNT
                  config: {}
      security:
        - ApiKeyAuth: []
components:
  parameters:
    market:
      name: market
      in: query
      description: Lend market identifier. Defaults to `main`.
      required: false
      style: form
      explode: true
      schema:
        $ref: '#/components/schemas/Market'
  schemas:
    BorrowPosition:
      required:
        - address
        - beforeBorrow
        - beforeSupply
        - borrow
        - borrowLiquidation
        - config
        - dustBorrow
        - id
        - isLiquidated
        - isSupplyPosition
        - ownerAddress
        - supply
        - supplyLiquidation
        - tick
        - tickId
        - vault
        - vaultId
      type: object
      properties:
        id:
          type: integer
          description: NFT / position id.
        vaultId:
          type: integer
        address:
          type: string
          description: Position NFT mint address.
        supply:
          type: string
        beforeSupply:
          type: string
        borrow:
          type: string
        beforeBorrow:
          type: string
        isLiquidated:
          type: boolean
        supplyLiquidation:
          type: string
        borrowLiquidation:
          type: string
        isSupplyPosition:
          type: boolean
        tick:
          type: integer
        tickId:
          type: integer
        dustBorrow:
          type: string
        ownerAddress:
          type: string
        vault:
          $ref: '#/components/schemas/BorrowVault'
        config:
          type: object
          additionalProperties:
            type: string
          description: Key/value position config from the database, keyed in camelCase.
      description: >-
        Serialized position data from `getVaultPositionsByUser()` in
        `app/core/borrowing.ts`.
    Market:
      type: string
      default: main
      enum:
        - main
        - ethena
    BorrowVault:
      required:
        - address
        - borrowToken
        - id
        - metadata
        - supplyToken
      type: object
      properties:
        id:
          type: integer
        address:
          type: string
        supplyToken:
          $ref: '#/components/schemas/Token'
        borrowToken:
          $ref: '#/components/schemas/Token'
        totalSupply:
          type: string
        totalSupplyLiquidity:
          type: string
        totalBorrow:
          type: string
        totalBorrowLiquidity:
          type: string
        absorbedSupply:
          type: string
        absorbedBorrow:
          type: string
        supplyRateMagnifier:
          type: string
        borrowRateMagnifier:
          type: string
        borrowFee:
          type: string
        collateralFactor:
          type: string
        liquidationThreshold:
          type: string
        liquidationMaxLimit:
          type: string
        liquidationPenalty:
          type: string
        withdrawalGap:
          type: string
        supplyRate:
          type: string
        supplyRateLiquidity:
          type: string
        borrowRate:
          type: string
        borrowRateLiquidity:
          type: string
        withdrawLimit:
          type: string
        withdrawableUntilLimit:
          type: string
        withdrawable:
          type: string
        borrowLimit:
          type: string
        borrowableUntilLimit:
          type: string
        borrowable:
          type: string
        borrowLimitUtilization:
          type: string
        minimumBorrowing:
          type: string
        totalPositions:
          type: integer
        oracle:
          type: string
        oraclePrice:
          type: string
        oraclePriceOperate:
          type: string
        oraclePriceLiquidate:
          type: string
        oracleSources:
          type: array
          items:
            $ref: '#/components/schemas/OracleSource'
        oracleTimestamp:
          type: integer
          format: int64
        liquiditySupplyData:
          $ref: '#/components/schemas/LiquiditySupplyData'
        liquidityBorrowData:
          $ref: '#/components/schemas/LiquidityBorrowData'
        rewards:
          type: array
          items:
            $ref: '#/components/schemas/MarketReward'
        topTick:
          type: integer
        metadata:
          $ref: '#/components/schemas/VaultMetadata'
      description: >-
        Serialized vault data from `serializedVault()` in
        `app/core/borrowing.ts`.
    Token:
      required:
        - address
        - chainId
        - decimals
        - name
        - price
        - symbol
        - uiSymbol
      type: object
      properties:
        address:
          type: string
          description: SPL token mint address.
        chainId:
          type: string
          example: solana
        name:
          type: string
        symbol:
          type: string
        uiSymbol:
          type: string
        decimals:
          type: integer
        logoUrl:
          type: string
          format: uri
        price:
          type: string
          description: Token price as a decimal string.
        coingeckoId:
          type: string
        updatedAt:
          type: string
          format: date-time
        stakingApr:
          type: number
          description: Present on vault supply/borrow tokens when applicable.
        nativeStakeVoteAccount:
          type: string
          description: Present on native-stake collateral tokens when applicable.
        scaledUiConfig:
          type: object
          description: >-
            Token-2022 scaled-UI-amount config. Present on scaled-UI tokens
            (e.g. tokenised stocks).
          properties:
            authority:
              type: string
            multiplier:
              type: number
            newMultiplier:
              type: number
            newMultiplierEffectiveTimestamp:
              type: string
    OracleSource:
      required:
        - divisor
        - invert
        - multiplier
        - source
        - sourceType
      type: object
      properties:
        source:
          type: string
        invert:
          type: boolean
        multiplier:
          type: string
        divisor:
          type: string
        sourceType:
          type: object
          additionalProperties: true
          example:
            chainlink: {}
    LiquiditySupplyData:
      type: object
      properties:
        modeWithInterest:
          type: boolean
        supply:
          type: string
        withdrawalLimit:
          type: string
        lastUpdateTimestamp:
          type: string
        expandPercent:
          type: integer
        expandDuration:
          oneOf:
            - type: string
            - type: integer
        baseWithdrawalLimit:
          type: string
        withdrawableUntilLimit:
          type: string
        withdrawable:
          type: string
    LiquidityBorrowData:
      type: object
      properties:
        modeWithInterest:
          type: boolean
        borrow:
          type: string
        borrowLimit:
          type: string
        lastUpdateTimestamp:
          oneOf:
            - type: string
            - type: integer
        expandPercent:
          type: integer
        expandDuration:
          type: integer
        baseBorrowLimit:
          type: string
        maxBorrowLimit:
          type: string
        borrowLimitUtilization:
          type: string
        borrowableUntilLimit:
          type: string
        borrowable:
          type: string
    MarketReward:
      required:
        - apr
        - metadata
        - rewardToken
        - side
        - type
      type: object
      properties:
        rewardToken:
          $ref: '#/components/schemas/Token'
        type:
          type: string
          enum:
            - merkle
        side:
          type: string
          enum:
            - supply
            - borrow
        apr:
          type: string
        metadata:
          type: object
          additionalProperties: true
    VaultMetadata:
      required:
        - maxLeverage
        - multiply
        - nativeStake
        - pegged
      type: object
      properties:
        maxLeverage:
          type: boolean
        nativeStake:
          type: boolean
        pegged:
          type: boolean
        multiply:
          required:
            - enabled
            - reduceFactor
          type: object
          properties:
            enabled:
              type: boolean
            reduceFactor:
              type: integer
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.