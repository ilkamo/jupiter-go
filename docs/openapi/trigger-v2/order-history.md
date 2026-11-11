> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Order History (Trigger API)

> Get historical trigger orders for an account



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml get /orders/history
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
  /orders/history:
    get:
      summary: Get order history
      description: |
        Retrieve paginated order history for the authenticated wallet.
        Filter by state, mint, and sort order.
      parameters:
        - name: state
          in: query
          schema:
            type: string
            enum:
              - active
              - past
          description: Filter by order state group
        - name: mint
          in: query
          schema:
            type: string
          description: Filter by token mint address
        - name: limit
          in: query
          schema:
            type: number
            default: 20
            minimum: 1
            maximum: 100
          description: Results per page
        - name: offset
          in: query
          schema:
            type: number
            default: 0
          description: Number of results to skip
        - name: sort
          in: query
          schema:
            type: string
            enum:
              - updated_at
              - created_at
              - expires_at
            default: updated_at
          description: Sort field
        - name: dir
          in: query
          schema:
            type: string
            enum:
              - asc
              - desc
            default: desc
          description: Sort direction
      responses:
        '200':
          description: Paginated order history
          content:
            application/json:
              schema:
                type: object
                properties:
                  orders:
                    type: array
                    items:
                      $ref: '#/components/schemas/OrderHistoryItem'
                  pagination:
                    type: object
                    properties:
                      total:
                        type: number
                      limit:
                        type: number
                      offset:
                        type: number
                required:
                  - orders
                  - pagination
        '401':
          description: Unauthorized
        '403':
          description: Forbidden
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    OrderHistoryItem:
      type: object
      properties:
        id:
          type: string
        orderType:
          type: string
          enum:
            - single
            - oco
            - otoco
        orderState:
          type: string
          enum:
            - pending
            - open
            - executing
            - filled
            - pending_withdraw
            - cancelled
            - expired
            - failed
        rawState:
          type: string
        userPubkey:
          type: string
        privyWalletPubkey:
          type: string
        inputMint:
          type: string
        initialInputAmount:
          type: string
        remainingInputAmount:
          type: string
        outputMint:
          type: string
        triggerMint:
          type: string
        triggerCondition:
          type: string
          enum:
            - above
            - below
        triggerPriceUsd:
          type: number
        slippageBps:
          type: number
        expiresAt:
          type: number
          nullable: true
        createdAt:
          type: number
        updatedAt:
          type: number
        triggeredAt:
          type: number
          nullable: true
        outputAmount:
          type: string
          nullable: true
        inputUsed:
          type: string
          nullable: true
        fillPercent:
          type: number
          nullable: true
        trailingBps:
          type: number
          nullable: true
          description: >-
            Trailing stop loss trail distance in basis points. Null for
            non-trailing orders.
        highWatermark:
          type: number
          nullable: true
          description: Highest observed trigger-mint price (sell-below trailing).
        lowWatermark:
          type: number
          nullable: true
          description: Lowest observed trigger-mint price (buy-above trailing).
        events:
          type: array
          items:
            type: object
            properties:
              type:
                type: string
                enum:
                  - deposit
                  - fill
                  - withdrawal
                  - cancelled
                  - expired
              timestamp:
                type: number
              txSignature:
                type: string
              state:
                type: string
                enum:
                  - success
                  - failed
                  - pending
              mint:
                type: string
                description: Token mint address (present on deposit/withdrawal events)
              amount:
                type: string
                description: >-
                  Token amount in smallest unit (present on deposit/withdrawal
                  events)
              outputMint:
                type: string
                description: Output token mint address (present on fill events)
              outputAmount:
                type: string
                description: Output token amount in smallest unit (present on fill events)
              orderContext:
                type: string
                description: Order context identifier (present on fill events)
      required:
        - id
        - orderType
        - orderState
        - userPubkey
        - inputMint
        - outputMint
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