> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Orderbook

> Get orderbook data for a specific market



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /orderbook/{marketId}
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
  /orderbook/{marketId}:
    get:
      tags:
        - Orderbook
      summary: Get orderbook
      description: >-
        Current orderbook depth for a market: YES and NO levels in contracts and
        in dollars.
      parameters:
        - schema:
            type: string
            minLength: 1
            description: Kalshi or Polymarket market identifier
            example: KXF1-25-LN
          required: true
          description: Kalshi or Polymarket market identifier
          name: marketId
          in: path
      responses:
        '200':
          description: Orderbook depth for the given market
          content:
            application/json:
              schema:
                type: object
                nullable: true
                properties:
                  'yes':
                    type: array
                    description: >-
                      YES-side levels sorted by price ascending. Each entry is a
                      [price_cents, size] tuple. price_cents is rounded, so
                      sub-cent levels show 0; size may be fractional.
                    items:
                      type: array
                      items:
                        type: number
                      minItems: 2
                      maxItems: 2
                  'no':
                    type: array
                    description: >-
                      NO-side levels sorted by price ascending. Each entry is a
                      [price_cents, size] tuple. price_cents is rounded, so
                      sub-cent levels show 0; size may be fractional.
                    items:
                      type: array
                      items:
                        type: number
                      minItems: 2
                      maxItems: 2
                  yes_dollars:
                    type: array
                    description: >-
                      YES-side levels with price as decimal dollar string. Each
                      entry is a [price_string, size] tuple, e.g. ["0.0100",
                      346014].
                    items:
                      type: array
                      minItems: 2
                      maxItems: 2
                  no_dollars:
                    type: array
                    description: >-
                      NO-side levels with price as decimal dollar string. Each
                      entry is a [price_string, size] tuple, e.g. ["0.0100",
                      219].
                    items:
                      type: array
                      minItems: 2
                      maxItems: 2
                required:
                  - 'yes'
                  - 'no'
                  - yes_dollars
                  - no_dollars
        '502':
          description: Failed to fetch orderbook data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    ErrorResponse:
      type: object
      properties:
        type:
          type: string
          enum:
            - invalid_request_error
            - authentication_error
            - permission_error
            - idempotency_error
            - rate_limit_error
            - api_error
        message:
          type: string
        code:
          type: string
        param:
          type: string
        request_id:
          type: string
        doc_url:
          type: string
      required:
        - type
        - message
        - request_id
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.