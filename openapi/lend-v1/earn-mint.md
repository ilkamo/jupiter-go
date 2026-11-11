> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Mint

> Request for a base64-encoded unsigned earn mint transaction to mint shares



## OpenAPI

````yaml openapi-spec/lend/lend.yaml post /earn/mint
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
  /earn/mint:
    post:
      summary: mint
      description: >
        Request for a base64-encoded unsigned earn mint transaction to mint
        shares
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/EarnSharesRequestBody'
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TransactionResponse'
              example:
                transaction: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
      security:
        - ApiKeyAuth: []
components:
  schemas:
    EarnSharesRequestBody:
      type: object
      required:
        - asset
        - signer
        - shares
      properties:
        asset:
          type: string
          default: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer:
          type: string
          default: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        shares:
          type: string
          default: '1000000'
      example:
        asset: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        shares: '1000000'
    TransactionResponse:
      type: object
      required:
        - transaction
      properties:
        transaction:
          type: string
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.