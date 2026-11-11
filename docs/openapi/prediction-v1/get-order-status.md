> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Order Status

> Get the latest status and history for an order



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /orders/status/{orderPubkey}
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
  /orders/status/{orderPubkey}:
    get:
      tags:
        - Orders
      summary: Get order status
      description: >-
        Latest status of an order with its event history. Poll it after `POST
        /execute` to follow the fill.
      parameters:
        - schema:
            type: string
            minLength: 1
            description: Order account public key
          required: true
          description: Order account public key
          name: orderPubkey
          in: path
      responses:
        '200':
          description: Latest status for an order with event history
          content:
            application/json:
              schema:
                type: object
                properties:
                  orderPubkey:
                    type: string
                  status:
                    type: string
                  latestEventType:
                    type: string
                  latestSignature:
                    type: string
                  externalOrderId:
                    type: string
                  orderId:
                    type: string
                  history:
                    type: array
                    items:
                      type: object
                      properties:
                        eventType:
                          type: string
                        status:
                          type: string
                        rawStatus:
                          type: string
                        timestamp:
                          type: integer
                        signature:
                          type: string
                        externalOrderId:
                          type: string
                        orderId:
                          type: string
                      required:
                        - eventType
                        - status
                        - rawStatus
                        - timestamp
                        - signature
                        - externalOrderId
                        - orderId
                required:
                  - orderPubkey
                  - status
                  - latestEventType
                  - latestSignature
                  - externalOrderId
                  - orderId
                  - history
        '400':
          description: Failed to fetch order status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Order history not found
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