> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Trades

> Get recent filled trades across all markets



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /trades
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
  /trades:
    get:
      tags:
        - Trades
      summary: Get trade feed
      description: Recent fills (`order_filled` events) across the exchange.
      responses:
        '200':
          description: Recent order_filled events
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: string
                          description: Trade identifier (e.g. "order-1906029")
                        ownerPubkey:
                          type: string
                        marketId:
                          type: string
                        message:
                          type: string
                        timestamp:
                          type: number
                        action:
                          type: string
                          enum:
                            - buy
                            - sell
                        side:
                          type: string
                          enum:
                            - 'yes'
                            - 'no'
                        eventTitle:
                          type: string
                        marketTitle:
                          type: string
                        amountUsd:
                          type: string
                        priceUsd:
                          type: string
                        eventImageUrl:
                          type: string
                        eventId:
                          type: string
                        isTeamMarket:
                          type: boolean
                        marketOptions:
                          type: array
                          items:
                            type: object
                            properties:
                              label:
                                type: string
                              buyYes:
                                type: boolean
                      required:
                        - id
                        - ownerPubkey
                        - marketId
                        - message
                        - timestamp
                        - action
                        - side
                        - eventTitle
                        - marketTitle
                        - amountUsd
                        - priceUsd
                        - eventImageUrl
                        - eventId
                required:
                  - data
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.