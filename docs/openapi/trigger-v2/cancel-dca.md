> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Cancel DCA Order

> Initiate cancellation of a DCA order and craft the withdrawal transaction



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /orders/dca/cancel/{id}
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
  /orders/dca/cancel/{id}:
    post:
      summary: Initiate DCA cancellation
      description: >
        Step 1 of the two-step cancellation. Moves the order to `withdrawing`
        and returns an

        unsigned withdrawal transaction for the unfilled remainder.
        Already-filled rounds are

        not reversed.
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: DCA order UUID
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
                  roundsRemaining:
                    type: number
                    description: Rounds that will not execute
                  refundAmount:
                    type: string
                    description: Amount to be refunded in smallest units
                  transaction:
                    type: string
                    description: Base64-encoded unsigned withdrawal transaction
                  requestId:
                    type: string
                    description: Pass to confirm-cancel as cancelRequestId
                required:
                  - id
                  - roundsRemaining
                  - refundAmount
                  - transaction
                  - requestId
              example:
                id: 019f0530-36a7-77da-a38f-ed9414bb996b
                roundsRemaining: 2
                refundAmount: '20000000'
                transaction: >-
                  AwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMBAgdyOCX48yeQcZusgzCLBJ+otTpXtR9YKzHDx3aLqYwC+0vQ4YrooVs15oSv6alpvSIyPs+Vo55usVe/6EytcDKmDUjHKsrOFPVwGfmQFwHHgWST75hdcLkGE1QZ/OgrO9V7X8BkIOtYGIrydcacmuIKcdfd6e/c7eONpxrCoSbqmM5Q+ul2G2Ootg7YYGSKJ3qW1hW2uKHAvmby4AdEXypFAwZGb+UhFzL/7K26csOb57yM5bvF9xJrLEObOkAAAAAG3fbh12Whk9nL4UbO63msHLSF7V9bN5E6jPWFfv8AqSI5PGNkRpV42tULB0GGLK8o3q7WErcVy5goaCJGrByHBAUABQJQAgAABQAJA5DQAwAAAAAABgQDBAECCQMALTEBAAAAAAYDAwABAQk=
                requestId: 368645e8-7ac6-4c88-8b10-3a68812bdc1d
        '400':
          description: Order not in active or withdrawing state, or no remaining funds
        '404':
          description: Order not found or not owned by the authenticated wallet
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