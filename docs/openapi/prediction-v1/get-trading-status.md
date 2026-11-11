> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Trading Status

> Get current trading status of the prediction market



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /trading-status
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
  /trading-status:
    get:
      tags:
        - Trading
      summary: Get trading status
      description: >-
        Whether the exchange currently accepts orders (`trading_active`). Check
        it before building an order.
      responses:
        '200':
          description: Whether trading is currently enabled
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TradingStatusResponse'
components:
  schemas:
    TradingStatusResponse:
      type: object
      properties:
        trading_active:
          type: boolean
      required:
        - trading_active
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.