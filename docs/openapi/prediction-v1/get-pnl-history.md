> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get PnL History

> Get historical PnL data for charting



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /profiles/{ownerPubkey}/pnl-history
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
  /profiles/{ownerPubkey}/pnl-history:
    get:
      tags:
        - Profile
      summary: Get P&L history
      description: >-
        Profit and loss time series of a wallet for charting: `count` points
        spaced by `interval`.
      parameters:
        - schema:
            type: string
          required: true
          description: Wallet public key.
          name: ownerPubkey
          in: path
        - schema:
            type: string
            enum:
              - 24h
              - 1w
              - 1m
            default: 1w
          required: false
          description: 'Spacing between points: `24h`, `1w` or `1m`.'
          name: interval
          in: query
        - schema:
            type: integer
            minimum: 0
            exclusiveMinimum: true
            maximum: 1000
            default: 10
          required: false
          description: Number of points to return.
          name: count
          in: query
      responses:
        '200':
          description: PnL history for charting
          content:
            application/json:
              schema:
                type: object
                properties:
                  ownerPubkey:
                    type: string
                  history:
                    type: array
                    items:
                      type: object
                      properties:
                        timestamp:
                          type: integer
                        realizedPnlUsd:
                          type: string
                      required:
                        - timestamp
                        - realizedPnlUsd
                required:
                  - ownerPubkey
                  - history
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.