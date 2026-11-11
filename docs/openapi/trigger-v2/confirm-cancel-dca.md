> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Confirm Cancel DCA Order

> Submit the signed withdrawal transaction to finalize a DCA cancellation



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/dca/confirm-cancel/{id}
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
  /orders/dca/confirm-cancel/{id}:
    post:
      summary: Confirm DCA cancellation
      description: >
        Step 2 of the two-step cancellation. Submit the signed withdrawal
        transaction to return

        the remaining funds and move the order to `cancelled`.
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: DCA order UUID
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
                  description: requestId from the cancel response
              required:
                - signedTransaction
                - cancelRequestId
      responses:
        '200':
          description: Cancellation confirmed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TxSignatureResponse'
              example:
                id: 019f0530-36a7-77da-a38f-ed9414bb996b
                txSignature: >-
                  53CQkssz7oGmbUxp3W5tsjw8Vhpv8FKVXjfWb3ZyPW71gsAm4qHvK2FjEDKt9RNDVGesoNiMA7moGR1rKbzhEfqb
        '400':
          description: >-
            Order not in withdrawing state, transaction mismatch, or invalid
            signature
        '401':
          description: Unauthorized
        '404':
          description: Order not found or not owned by the authenticated wallet
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