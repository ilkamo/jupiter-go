> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Vault Info

> Get vault account information and balance



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /vault-info
openapi: 3.0.3
info:
  title: Jupiter Prediction Market API
  version: 1.0.0
  description: Jupiter Prediction Market API Schema
servers:
  - url: https://api.jup.ag/prediction/v1
    description: Jupiter Prediction Market API Endpoint
security:
  - ApiKeyAuth: []
tags:
  - name: Events
  - name: Markets
  - name: Orders
  - name: Positions
  - name: History
  - name: Trading
  - name: Orderbook
  - name: Milestones
  - name: Vault
paths:
  /vault-info:
    get:
      tags:
        - Vault
      summary: Get vault info
      description: 'Settlement vault account: public key, decoded account data and balance.'
      responses:
        '200':
          description: Vault account data
          content:
            application/json:
              schema:
                type: object
                properties:
                  pubkey:
                    type: string
                  data:
                    type: object
                    additionalProperties:
                      type: string
                  vaultBalance:
                    type: string
                required:
                  - pubkey
                  - data
                  - vaultBalance
        '404':
          description: Vault account not found
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.