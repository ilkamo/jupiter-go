> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# DBC Fee Create TX

> Request for a base64-encoded unsigned transaction to claim creator trading fees of a Dynamic Bonding Curve pool



## OpenAPI

````yaml openapi-spec/studio/studio.yaml post /dbc/fee/create-tx
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
  /dbc/fee/create-tx:
    post:
      summary: dbc-fee-create-tx
      description: >
        Request for a base64-encoded unsigned transaction to claim creator
        trading fees of a Dynamic Bonding Curve pool

        - Handles both direct creator ownership and proxy-based ownership.
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateClaimFeeDBCTransactionRequestBody'
            example:
              ownerWallet: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              poolAddress: PoolAddress11111111111111111111111111111111
              maxQuoteAmount: 5000000
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                required:
                  - transaction
                properties:
                  transaction:
                    type: string
                    description: Base64-encoded transaction ready for signing
              example:
                transaction: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  errors:
                    type: object
                    additionalProperties:
                      type: array
                      items:
                        type: string
        '403':
          description: Not authorized to claim fees from this pool
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    example: You are not the owner of this pool
        '404':
          description: Proxy account not found for the creator
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    example: Proxy account not found for the creator
        '500':
          description: Internal server error
      security:
        - ApiKeyAuth: []
components:
  schemas:
    CreateClaimFeeDBCTransactionRequestBody:
      type: object
      required:
        - ownerWallet
        - poolAddress
        - maxQuoteAmount
      properties:
        ownerWallet:
          type: string
          description: |
            - Owner wallet public key
        poolAddress:
          type: string
          description: |
            - Dynamic Bonding Curve pool address
        maxQuoteAmount:
          type: number
          description: |
            - Maximum quote amount to claim
          minimum: 0
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.