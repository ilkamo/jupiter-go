> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# DBC Pool Create TX

> Request for a base64-encoded unsigned transaction to create a Dynamic Bonding Curve pool with token metadata



## OpenAPI

````yaml openapi-spec/studio/studio.yaml post /dbc-pool/create-tx
openapi: 3.0.3
info:
  title: Jupiter Studio API
  version: 2.0.0
  description: Jupiter Studio API Schema
servers:
  - url: https://api.jup.ag/studio/v1
    description: Jupiter Studio API Endpoint
security: []
paths:
  /dbc-pool/create-tx:
    post:
      summary: dbc-pool-create-tx
      description: >
        Request for a base64-encoded unsigned transaction to create a Dynamic
        Bonding Curve pool with token metadata
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateDBCTransactionRequestBody'
            example:
              buildCurveByMarketCapParam:
                quoteMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
                initialMarketCap: 16000
                migrationMarketCap: 69000
                tokenQuoteDecimal: 6
                lockedVestingParam:
                  totalLockedVestingAmount: 0
                  cliffUnlockAmount: 0
                  numberOfVestingPeriod: 0
                  totalVestingDuration: 0
                  cliffDurationFromMigrationTime: 0
              antiSniping: true
              fee:
                feeBps: 100
              isLpLocked: true
              tokenName: My Token
              tokenSymbol: MTK
              tokenImageContentType: image/png
              creator: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateDBCTransactionResponse'
              example:
                transaction: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
                mint: TokenMintAddressExample111111111111111111111
                imagePresignedUrl: https://storage.example.com/upload/image?token=abc123
                metadataPresignedUrl: https://storage.example.com/upload/metadata?token=def456
                imageUrl: >-
                  https://storage.example.com/tokens/TokenMintAddressExample111111111111111111111/image.png
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  errors:
                    type: object
                  error:
                    type: string
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                  details:
                    type: string
      security:
        - ApiKeyAuth: []
components:
  schemas:
    CreateDBCTransactionRequestBody:
      type: object
      required:
        - buildCurveByMarketCapParam
        - antiSniping
        - fee
        - tokenName
        - tokenSymbol
        - tokenImageContentType
        - creator
      properties:
        buildCurveByMarketCapParam:
          type: object
          required:
            - quoteMint
            - initialMarketCap
            - migrationMarketCap
            - tokenQuoteDecimal
            - lockedVestingParam
          properties:
            quoteMint:
              type: string
              description: |
                - Quote mint address
                - Either USDC, SOL, or JUP
              example: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
            initialMarketCap:
              type: number
              description: |
                - Initial market cap in quote token units
                - 16000 passed in = 16000 USDC/SOL/JUP initial market cap
              example: 16000
            migrationMarketCap:
              type: number
              description: |
                - Migration market cap in quote token units
                - 69000 passed in = 69000 USDC/SOL/JUP migration market cap
              example: 69000
            tokenQuoteDecimal:
              type: number
              enum:
                - 6
                - 9
              description: |
                - Token quote decimal places
            lockedVestingParam:
              type: object
              required:
                - totalLockedVestingAmount
                - cliffUnlockAmount
                - numberOfVestingPeriod
                - totalVestingDuration
                - cliffDurationFromMigrationTime
              properties:
                totalLockedVestingAmount:
                  type: number
                  description: >
                    - Total locked vesting amount of tokens (0% to 80% of total
                    supply)
                cliffUnlockAmount:
                  type: number
                  description: |
                    - Amount of tokens unlocked at cliff
                numberOfVestingPeriod:
                  type: number
                  enum:
                    - 0
                    - 6
                    - 12
                    - 183
                    - 365
                  description: >
                    - Number of vesting periods in days (183/365) or months
                    (6/12)
                totalVestingDuration:
                  type: number
                  enum:
                    - 0
                    - 15552000
                    - 15811200
                    - 31536000
                  description: |
                    - Total vesting duration in seconds
                cliffDurationFromMigrationTime:
                  type: number
                  enum:
                    - 0
                    - 15552000
                    - 15811200
                    - 31536000
                  description: |
                    - Cliff duration from migration time in seconds
        antiSniping:
          type: boolean
          description: >
            - Enable anti-sniping protection

            - Apply an additional swap fee immediately after launch to
            discourage bots

            - Starts at 99% and decreases 1% linearly over a randomized period
        fee:
          type: object
          description: |
            - Fee configuration parameters
            - If not provided, the default fee will be 100 basis points (1%)
          properties:
            totalDuration:
              type: number
              enum:
                - 60
                - 120
                - 180
                - 240
                - 300
              description: Total duration for fee schedule
            baseFeeMode:
              type: string
              enum:
                - linear
                - exponential
              description: Base fee mode for fee scheduler
            feeBps:
              type: number
              enum:
                - 100
                - 200
              description: Ending fee in basis points (100 or 200)
              default: 100
          required:
            - feeBps
        isLpLocked:
          type: boolean
          description: |
            - Whether LP tokens should be locked
            - If enabled, 50% of graduated LP unlocks after 1 year
            - Useful for creators to strategize in the long term
          default: true
        tokenName:
          type: string
          minLength: 1
        tokenSymbol:
          type: string
          minLength: 1
        tokenImageContentType:
          type: string
          description: |
            - Token image content type
          enum:
            - image/jpeg
            - image/png
            - image/gif
            - image/webp
        creator:
          type: string
          description: |
            - Creator wallet public key
    CreateDBCTransactionResponse:
      type: object
      required:
        - transaction
        - mint
        - imagePresignedUrl
        - metadataPresignedUrl
        - imageUrl
      properties:
        transaction:
          type: string
          description: Base64-encoded transaction ready for signing
        mint:
          type: string
          description: Generated token mint address
        imagePresignedUrl:
          type: string
          format: uri
          description: >
            - Presigned URL for image upload via `PUT` request

            - This is for you to make a PUT request to upload your token image
            to the on-chain uri metadata

            - [Refer to the Token Metadata section for better
            understanding](/docs/studio/create-token#token-metadata)
        metadataPresignedUrl:
          type: string
          format: uri
          description: >
            - Presigned URL for metadata upload via `PUT` request

            - This is for you to make a PUT request to upload your token
            metadata to the on-chain uri metadata

            - [Refer to the Token Metadata section for better
            understanding](/docs/studio/create-token#token-metadata)
        imageUrl:
          type: string
          format: uri
          description: Final image URL
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.