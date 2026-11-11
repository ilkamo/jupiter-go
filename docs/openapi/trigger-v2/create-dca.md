> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create DCA Order

> Create a recurring dollar-cost averaging order



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/dca
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
  /orders/dca:
    post:
      summary: Create DCA order
      description: >
        Create a dollar-cost averaging order that splits a single deposit into
        `orderCount`

        equal rounds executed every `intervalSeconds`. Submit the signed deposit
        transaction

        from `/deposit/craft` (crafted with `orderType: "dca"`).


        Call this path with no trailing slash. `POST /orders/dca/` returns 404.


        Rounds execute as Jupiter swaps with Ultra-managed (RTSE) slippage;
        there is no

        slippage parameter. Each round pays the standard Jupiter swap fee (no
        separate DCA fee),

        and the output settles directly to the taker's wallet.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                depositRequestId:
                  type: string
                  description: requestId from the deposit craft response
                depositSignedTx:
                  type: string
                  description: Base64-encoded signed deposit transaction
                userPubkey:
                  type: string
                  description: Your wallet public key (must match the JWT)
                inputMint:
                  type: string
                  description: Mint of the token to sell
                outputMint:
                  type: string
                  description: Mint of the token to buy
                inputAmount:
                  type: string
                  description: >-
                    Total amount to DCA in smallest units. Dust from uneven
                    division is added to the last round.
                orderCount:
                  type: number
                  minimum: 2
                  description: >-
                    Number of rounds to split the deposit into (minimum 2). No
                    fixed upper limit, but each round must be worth at least 10
                    USD, so the practical maximum is floor(total deposit USD /
                    10).
                intervalSeconds:
                  type: number
                  minimum: 60
                  maximum: 31536000
                  description: >-
                    Seconds between rounds (60 to 31,536,000, i.e. 1 minute to 1
                    year)
                orderType:
                  type: string
                  enum:
                    - time_based
                    - price_conditional
                  default: time_based
                  description: >-
                    time_based runs every interval; price_conditional only fills
                    while the trigger mint price is within the band.
                minPriceUsd:
                  type: number
                  description: Lower USD price bound. Price-conditional only.
                maxPriceUsd:
                  type: number
                  description: Upper USD price bound. Price-conditional only.
                triggerMint:
                  type: string
                  description: >-
                    Mint whose USD price is checked against the band. Required
                    for price_conditional. Must be a mint with a supported USD
                    price (in practice the volatile leg / outputMint); a
                    stablecoin such as USDC returns "Trigger mint is not
                    supported". Ignored for time_based.
                beginFillAt:
                  type: string
                  description: >-
                    ISO-8601 time to start the first round. Defaults to now, max
                    30 days in the future.
                jlEnabled:
                  type: boolean
                  default: false
                  description: >-
                    Earn While You Wait. When true, idle funds earn Jupiter Lend
                    yield between rounds. Requires orderType time_based, a
                    supported stablecoin input, and a deposit crafted with
                    jlMint. Default false.
                jlMint:
                  type: string
                  description: >-
                    Jupiter Lend earn token for the input stablecoin (from GET
                    /lend/v1/earn/tokens, the address whose assetAddress is the
                    inputMint). Required when jlEnabled is true.
              required:
                - depositRequestId
                - depositSignedTx
                - userPubkey
                - inputMint
                - outputMint
                - inputAmount
                - orderCount
                - intervalSeconds
            example:
              depositRequestId: a8551844-fa46-41c7-af3a-70340c465bde
              depositSignedTx: AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
              userPubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              inputMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
              outputMint: So11111111111111111111111111111111111111112
              inputAmount: '20000000'
              orderCount: 2
              intervalSeconds: 3600
              orderType: time_based
      responses:
        '200':
          description: DCA order created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TxSignatureResponse'
              example:
                id: 019f0530-36a7-77da-a38f-ed9414bb996b
                txSignature: >-
                  PPbp1QbbKDPjpHM3ScKyZBnCtLKKqmWwDeVj7USdBeEoNQsA1hkN35UQfrY2cQpyAB1pfVi9FdfciYdz7qUXBQA
        '400':
          description: >-
            Validation error, below minimum (10 USD per round), the max of 10
            active DCA orders reached ("Maximum of 10 active DCA orders allowed
            per user"), or invalid deposit
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Unauthorized
        '403':
          description: userPubkey does not match the authenticated wallet
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    TxSignatureResponse:
      type: object
      description: >-
        Minimal create/cancel response carrying the order ID and an on-chain
        transaction signature. Returned by DCA create (deposit signature) and by
        confirm-cancel (withdrawal signature). Unlike price order create, it
        does not include depositConfirmed.
      properties:
        id:
          type: string
          description: Order UUID
        txSignature:
          type: string
          description: On-chain transaction signature
      required:
        - id
        - txSignature
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