> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Update Order

> Update parameters of an existing trigger order



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml patch /orders/price/{orderId}
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
  /orders/price/{orderId}:
    patch:
      summary: Update price order
      description: Update trigger price or slippage of an existing order.
      parameters:
        - name: orderId
          in: path
          required: true
          schema:
            type: string
          description: Order UUID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                orderType:
                  type: string
                  enum:
                    - single
                    - oco
                    - otoco
                triggerPriceUsd:
                  type: number
                  description: >-
                    New trigger price. Cannot be set on a trailing stop loss
                    order.
                trailingBps:
                  type: number
                  minimum: 50
                  maximum: 9000
                  description: >-
                    New trailing stop loss trail distance in basis points
                    (50-9000), for an order that is already trailing.
                slippageBps:
                  type: number
                tpPriceUsd:
                  type: number
                slPriceUsd:
                  type: number
                tpSlippageBps:
                  type: number
                slSlippageBps:
                  type: number
              required:
                - orderType
            example:
              orderType: single
              triggerPriceUsd: 210
              slippageBps: 150
      responses:
        '200':
          description: Order updated
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                required:
                  - id
              example:
                id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
        '400':
          description: Validation error
        '401':
          description: Unauthorized
        '403':
          description: Order belongs to a different wallet
        '404':
          description: Order not found
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
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