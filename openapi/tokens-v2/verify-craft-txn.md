> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Transaction

> Craft an unsigned 1000 JUP payment transaction for express verification



## OpenAPI

````yaml openapi-spec/tokens/v2/verification.yaml get /express/craft-txn
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
  /express/craft-txn:
    get:
      tags:
        - Express Verification
      summary: Craft express verification payment transaction
      description: >
        Returns an unsigned Solana transaction worth 1000 JUP as payment for
        express

        verification. JUP transfers directly; SOL, USDC, and JUPUSD craft an
        Ultra swap to

        1000 JUP. The `requestId` in the response is required for the execute
        step.
      parameters:
        - in: query
          name: senderAddress
          schema:
            type: string
          required: true
          description: Solana wallet address that will sign and pay for the transaction
        - in: query
          name: paymentCurrency
          schema:
            $ref: '#/components/schemas/PaymentCurrency'
          required: false
          description: >-
            Payment currency. Defaults to JUP (direct transfer). SOL/USDC/JUPUSD
            route through an Ultra swap to JUP.
      responses:
        '200':
          description: Unsigned payment transaction
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CraftTxnResponse'
        '400':
          description: Invalid request parameters
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Failed to craft transaction
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
components:
  schemas:
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
    CraftTxnResponse:
      type: object
      required:
        - receiverAddress
        - mint
        - amount
        - tokenDecimals
        - feeLamports
        - feeMint
        - feeTokenDecimals
        - feeAmount
        - transaction
        - requestId
        - totalTime
        - code
        - gasless
      properties:
        transaction:
          type: string
          description: Base64-encoded unsigned transaction
        requestId:
          type: string
          description: Unique request ID. Pass this to the execute endpoint.
        receiverAddress:
          type: string
          description: Wallet address receiving the JUP payment
        mint:
          type: string
          description: Mint address of the payment token (JUP)
          example: JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN
        amount:
          type: string
          description: >-
            JUP amount collected in smallest unit. JUP path transfers exactly
            1000 JUP (1000000000); swap paths yield at least that.
        tokenDecimals:
          type: number
          description: Decimals of the payment token
        tokenUsdRate:
          type: number
          description: USD rate of the payment token at time of crafting
        feeLamports:
          type: number
          description: Transaction fee in lamports
        feeUsdAmount:
          type: number
          description: Transaction fee in USD
        feeMint:
          type: string
          description: Mint address of the fee token
        feeTokenDecimals:
          type: number
          description: Decimals of the fee token
        feeAmount:
          type: number
          description: Fee amount in smallest unit
        totalTime:
          type: number
          description: Time taken to craft the transaction in milliseconds
        expireAt:
          type: string
          description: Expiry timestamp for the crafted transaction
        error:
          type: string
          description: Error message if crafting failed
        code:
          type: number
          description: Status code (0 for success)
        gasless:
          type: boolean
          description: Whether the transaction is gasless
        inputMint:
          type: string
          description: Input token mint address (non-JUP paths)
        inputDecimals:
          type: number
          description: Input token decimals (non-JUP paths)
        quotedInputAmount:
          type: string
          description: >-
            Atomic input amount for the swap, sized to yield at least 1000 JUP
            (non-JUP paths)
        maxInputAmount:
          type: string
          description: >-
            Maximum atomic input amount for the swap; equal to quotedInputAmount
            on the current swap flow (non-JUP paths)
    ErrorResponse:
      type: object
      properties:
        status:
          type: number
        message:
          type: string
        error:
          type: string
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.