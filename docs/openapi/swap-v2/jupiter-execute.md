> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Execute Transaction

> Execute a signed transaction from /order with managed landing



## OpenAPI

````yaml openapi-spec/swap/v2/swap.yaml post /execute
openapi: 3.0.3
info:
  title: Swap API V2
  version: 2.0.0
  description: >
    Jupiter Swap API V2. Three endpoints: `/order` (default happy path with
    managed execution),

    `/build` (advanced path returning raw swap instructions), and `/execute`
    (managed transaction execution for `/order`).
servers:
  - url: https://api.jup.ag/swap/v2
    description: Swap API V2 Endpoint
security: []
paths:
  /execute:
    post:
      tags:
        - Swap
      summary: Execute Transaction
      description: >
        Execute a signed transaction from `/order`. Handles transaction landing,
        confirmation, and retry.


        Only available for transactions obtained via `/order`. Transactions
        built via `/build` must be

        sent through your own RPC.
      operationId: PostExecute
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - signedTransaction
                - requestId
              properties:
                signedTransaction:
                  type: string
                  description: Base64-encoded signed transaction
                requestId:
                  type: string
                  description: The requestId from the /order response
                lastValidBlockHeight:
                  type: string
                  description: Optional block height for nonce validation
      responses:
        '200':
          description: Execution result
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    enum:
                      - Success
                      - Failed
                  signature:
                    type: string
                    description: Transaction signature
                  slot:
                    type: string
                    description: Confirmed slot
                  error:
                    type: string
                    description: Error message (present if Failed)
                  code:
                    type: number
                    description: >
                      - Error code. 0 = success.

                      - Ultra: -1 (missing cached order), -2 (invalid signed
                      tx), -3 (invalid message bytes).

                      - Aggregator: -1000 (failed to land), -1001 (unknown),
                      -1002 (invalid tx), -1003 (not fully signed), -1004
                      (invalid block height), -1005 (transaction expired before
                      it landed).

                      - RFQ: -2000 (failed to land), -2001 (unknown), -2002
                      (invalid payload), -2003 (quote expired), -2004 (swap
                      rejected).
                  totalInputAmount:
                    type: string
                    description: >-
                      Total input token amount deducted from the user's wallet,
                      including the fee if feeMint is the input mint
                  totalOutputAmount:
                    type: string
                    description: >-
                      Final output token amount reflected in the user's wallet,
                      after any fee collected in the output mint
                  inputAmountResult:
                    type: string
                    description: >-
                      Input amount that went into the swap route, after any fee
                      collected in the input mint
                  outputAmountResult:
                    type: string
                    description: >-
                      Output amount produced by the swap route, before any fee
                      collected in the output mint
                  swapEvents:
                    type: array
                    items:
                      $ref: '#/components/schemas/SwapEvent'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                  code:
                    type: number
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                type: object
                properties:
                  signature:
                    type: string
                  error:
                    type: string
      security:
        - ApiKeyAuth: []
components:
  schemas:
    SwapEvent:
      type: object
      properties:
        inputMint:
          type: string
        inputAmount:
          type: string
        outputMint:
          type: string
        outputAmount:
          type: string
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.