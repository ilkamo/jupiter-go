> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Cancel Order (Trigger API)

> Cancel a pending trigger order



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/price/cancel/{orderId}
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
  /orders/price/cancel/{orderId}:
    post:
      summary: Initiate order cancellation
      description: >
        Step 1 of the two-step cancellation flow. Returns a withdrawal
        transaction

        that must be signed and submitted via the confirm-cancel endpoint.
      parameters:
        - name: orderId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Withdrawal transaction for signing
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  transaction:
                    type: string
                    description: Base64-encoded unsigned withdrawal transaction
                  requestId:
                    type: string
                    description: Cancel request ID (used in confirm step)
                required:
                  - id
                  - transaction
                  - requestId
              example:
                id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
                transaction: AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
                requestId: cancel-request-uuid
        '400':
          description: Order not in a cancellable state
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