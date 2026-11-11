> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Execute Verification (Tokens API)

> Submit signed transaction with verification and metadata details



## OpenAPI

````yaml openapi-spec/tokens/v2/verification.yaml post /express/execute
openapi: 3.0.3
info:
  title: Jupiter Token Verification API
  version: 1.0.0
  description: API for express token verification and metadata updates via Jupiter VRFD
servers:
  - url: https://api.jup.ag/tokens/v2/verify
    description: Jupiter Token Verification API Endpoint
security: []
paths:
  /express/execute:
    post:
      tags:
        - Express Verification
      summary: Execute express verification
      description: >
        Submits the signed payment transaction, records the payment, creates a
        verification

        request, and optionally submits token metadata updates. Verification and
        metadata

        updates are reviewed independently.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ExpressExecuteBody'
      responses:
        '200':
          description: Submission result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExpressExecuteResponse'
              example:
                status: Success
                signature: 5UfDuX...
                totalTime: 1234
                verificationCreated: true
                metadataCreated: true
        '400':
          description: Invalid request body
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '409':
          description: Token not eligible for verification or metadata update
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Transaction execution or server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
components:
  schemas:
    ExpressExecuteBody:
      type: object
      required:
        - transaction
        - requestId
        - senderAddress
        - tokenId
        - twitterHandle
        - description
      properties:
        transaction:
          type: string
          description: Base64-encoded signed transaction from the craft-txn step
        requestId:
          type: string
          description: requestId returned by craft-txn
        senderAddress:
          type: string
          description: Solana wallet address of the sender
        tokenId:
          type: string
          description: Token mint address to verify
        twitterHandle:
          type: string
          description: Twitter/X handle of the token project
        senderTwitterHandle:
          type: string
          description: Twitter/X handle of the submitter (optional)
        description:
          type: string
          description: Reason for verification request
        tokenMetadata:
          $ref: '#/components/schemas/TokenMetadataInput'
        paymentCurrency:
          allOf:
            - $ref: '#/components/schemas/PaymentCurrency'
            - description: >-
                Payment currency. Defaults to JUP (direct transfer).
                SOL/USDC/JUPUSD execute the Ultra swap crafted in the previous
                step.
        paymentAmount:
          type: string
          description: Atomic input amount paid. Required when paymentCurrency is not JUP.
        jupOutputAmount:
          type: string
          description: >-
            Atomic JUP output from the craft quote (CraftTxnResponse.amount),
            echoed back so revenue tracks the actual JUP collected on swap
            paths.
    ExpressExecuteResponse:
      type: object
      required:
        - status
        - totalTime
        - verificationCreated
        - metadataCreated
      properties:
        status:
          type: string
          enum:
            - Success
            - Failed
          description: Whether the transaction was executed successfully
        signature:
          type: string
          description: Transaction signature on-chain (present on success)
        error:
          type: string
          description: Error message (present on failure)
        code:
          type: number
          description: Error code (present on failure)
        totalTime:
          type: number
          description: Time taken to execute in milliseconds
        verificationCreated:
          type: boolean
          description: Whether a verification request was created
        metadataCreated:
          type: boolean
          description: Whether a metadata update request was created
    ErrorResponse:
      type: object
      properties:
        status:
          type: number
        message:
          type: string
        error:
          type: string
    TokenMetadataInput:
      type: object
      required:
        - tokenId
      description: Optional token metadata to submit alongside verification
      properties:
        tokenId:
          type: string
          description: Token mint address
        icon:
          type: string
          nullable: true
          description: URL to token icon image
        name:
          type: string
          nullable: true
          description: Token display name
        symbol:
          type: string
          nullable: true
          description: Token ticker symbol
        website:
          type: string
          nullable: true
          description: Project website URL
        telegram:
          type: string
          nullable: true
          description: Telegram group URL
        twitter:
          type: string
          nullable: true
          description: Twitter/X profile URL
        twitterCommunity:
          type: string
          nullable: true
          description: Twitter/X community URL
        discord:
          type: string
          nullable: true
          description: Discord invite URL
        instagram:
          type: string
          nullable: true
          description: Instagram profile URL
        tiktok:
          type: string
          nullable: true
          description: TikTok profile URL
        circulatingSupply:
          type: string
          nullable: true
          description: Circulating supply value
        useCirculatingSupply:
          type: boolean
          nullable: true
          description: Whether to use the provided circulating supply
        tokenDescription:
          type: string
          nullable: true
          description: Short description of the token
        coingeckoCoinId:
          type: string
          nullable: true
          description: CoinGecko coin ID for price data
        useCoingeckoCoinId:
          type: boolean
          nullable: true
          description: Whether to use the provided CoinGecko ID
        circulatingSupplyUrl:
          type: string
          nullable: true
          description: URL to a live circulating supply endpoint
        useCirculatingSupplyUrl:
          type: boolean
          nullable: true
          description: Whether to use the provided circulating supply URL
        otherUrl:
          type: string
          nullable: true
          description: Any other relevant URL
    PaymentCurrency:
      type: string
      enum:
        - JUP
        - SOL
        - USDC
        - JUPUSD
      description: >
        Accepted payment currency, given as a token symbol (not a mint address):
        JUP, SOL, USDC, or JUPUSD.

        All input currencies are converted to 1000 JUP. JUP transfers directly;
        SOL, USDC, and JUPUSD are

        swapped to 1000 JUP via Ultra.
      example: JUP
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.