> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Withdraw

> Request for a base64-encoded unsigned earn withdraw transaction to withdraw assets



## OpenAPI

````yaml openapi-spec/lend/lend.yaml post /earn/withdraw
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
  /earn/withdraw:
    post:
      summary: withdraw
      description: >
        Request for a base64-encoded unsigned earn withdraw transaction to
        withdraw assets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/EarnAmountRequestBody'
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
    EarnAmountRequestBody:
      type: object
      required:
        - asset
        - signer
        - amount
      properties:
        asset:
          type: string
          default: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer:
          type: string
          default: vBXNsd5SRtTPpW7GWv3wREA6Ztm2jCWp5eqqTsVhyG5
        amount:
          type: string
          default: '5'
      example:
        asset: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        amount: '1000000'
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