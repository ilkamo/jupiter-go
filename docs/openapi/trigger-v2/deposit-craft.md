> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Deposit

> Craft a deposit transaction for the trigger vault



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /deposit/craft
openapi: 3.0.0
info:
  title: Jupiter Trigger Order API V2
  version: 2.0.0
  description: >-
    Jupiter Trigger Order API V2 with vault-based deposits, JWT authentication,
    and advanced order types.
servers:
  - url: https://api.jup.ag/trigger/v2
    description: Jupiter Trigger Order API V2 Endpoint
security: []
paths:
  /deposit/craft:
    post:
      summary: Craft deposit transaction
      description: >
        Build an unsigned transaction to deposit tokens from your wallet into
        your vault.

        The vault address is automatically resolved from your JWT token.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                inputMint:
                  type: string
                  description: Mint address of the token to deposit
                outputMint:
                  type: string
                  description: Mint address of the desired output token
                userAddress:
                  type: string
                  description: Your wallet public key
                amount:
                  type: string
                  description: Amount in smallest unit (e.g. lamports for SOL)
                orderType:
                  type: string
                  enum:
                    - price
                    - dca
                  description: >-
                    Order type for the deposit. Use price for price trigger
                    orders, or dca for dollar-cost averaging orders.
                orderSubType:
                  type: string
                  enum:
                    - single
                    - oco
                    - otoco
                  description: >-
                    Required when orderType is price (single for standalone
                    orders, oco for OCO orders, otoco for OTOCO orders). Must be
                    omitted when orderType is dca.
                jlMint:
                  type: string
                  description: >-
                    For Earn While You Wait DCA orders only. The Jupiter Lend
                    earn token for the input stablecoin (from GET
                    /lend/v1/earn/tokens). Crafts a deposit into Jupiter Lend;
                    the response returns jlTokenAccount instead of
                    inputTokenAccount.
              required:
                - inputMint
                - outputMint
                - userAddress
                - amount
                - orderType
            example:
              inputMint: So11111111111111111111111111111111111111112
              outputMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
              userAddress: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              amount: '1000000000'
              orderType: price
              orderSubType: single
      responses:
        '200':
          description: Unsigned deposit transaction
          content:
            application/json:
              schema:
                type: object
                properties:
                  transaction:
                    type: string
                    description: Base64-encoded unsigned transaction
                  requestId:
                    type: string
                    description: Deposit request ID (used when creating the order)
                  receiverAddress:
                    type: string
                    description: Vault public key receiving the deposit
                  mint:
                    type: string
                    description: Token mint being deposited
                  amount:
                    type: string
                    description: Deposit amount in smallest unit
                  tokenDecimals:
                    type: number
                    description: Token decimals
                  inputTokenAccount:
                    type: string
                    description: >-
                      Deterministic per-order input token account. Returned for
                      deposits crafted with an orderSubType.
                  outputTokenAccount:
                    type: string
                    description: >-
                      Deterministic per-order output token account. Returned
                      only when orderSubType is otoco.
                  jlTokenAccount:
                    type: string
                    description: >-
                      Jupiter Lend position account. Returned instead of
                      inputTokenAccount when the deposit is crafted with jlMint
                      (Earn While You Wait).
                required:
                  - transaction
                  - requestId
                  - receiverAddress
                  - mint
                  - amount
                  - tokenDecimals
              example:
                transaction: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
                requestId: 01234567-89ab-cdef-0123-456789abcdef
                receiverAddress: 7nE9GJoYHNmtaQvTQpota3KV2oz4pQ2dA6nvYK8EUJHV
                mint: So11111111111111111111111111111111111111112
                amount: '1000000000'
                tokenDecimals: 9
                inputTokenAccount: InputTokenAccountPubkey...
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Unauthorized
        '403':
          description: User address mismatch or no vault registered
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        details:
          type: object
          additionalProperties:
            type: string
      required:
        - error
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT token from the challenge-response auth flow

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.