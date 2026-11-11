> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Confirm Cancel

> Confirm and execute a pending order cancellation



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/price/confirm-cancel/{orderId}
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
  /orders/price/confirm-cancel/{orderId}:
    post:
      summary: Confirm order cancellation
      description: >
        Step 2 of the two-step cancellation flow. Submit the signed withdrawal
        transaction

        to complete the cancellation and return funds to your wallet.
      parameters:
        - name: orderId
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                signedTransaction:
                  type: string
                  description: Base64-encoded signed withdrawal transaction
                cancelRequestId:
                  type: string
                  description: Cancel request ID from the initiate step
              required:
                - signedTransaction
                - cancelRequestId
            example:
              signedTransaction: AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
              cancelRequestId: cancel-request-uuid
      responses:
        '200':
          description: Cancellation confirmed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TxSignatureResponse'
              example:
                id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
                txSignature: 3PDo3pMJLqvAfNXwYjY7BSP57ZNQ5DgYDeK6xYszUVMn...
        '400':
          description: Invalid transaction or request ID
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
  schemas:
    TxSignatureResponse:
      type: object
      description: >-
        Minimal create/cancel response carrying the order ID and an on-chain
        transaction signature. Returned by DCA create (deposit signature) and by
        confirm-cancel (withdrawal signature). Unlike price order create, it
        does not include depositConfirmed.
      properties:
        id:
          type: string
          description: Order UUID
        txSignature:
          type: string
          description: On-chain transaction signature
      required:
        - id
        - txSignature
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