> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Execute Order (Prediction API)

> Submit a signed order transaction, including Jupiter Forecast orders



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml post /execute
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
  /execute:
    post:
      tags:
        - Orders
      summary: Execute a signed order transaction
      description: >-
        Submit a signed order transaction for execution. Pass the base64
        `signedTransaction` together with the `execution.context` object
        returned by the order build (`POST /orders`, or `DELETE
        /positions/{positionPubkey}` for a sell), unchanged. Jupiter Forecast
        (bisonfi) orders execute as an atomic swap through this endpoint and the
        transaction signature is returned on success.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ExecuteRequest'
      responses:
        '200':
          description: Execution result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExecuteResponse'
        '400':
          description: Execution failed or invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    ExecuteRequest:
      type: object
      properties:
        signedTransaction:
          type: string
          description: Base64-encoded signed transaction from the order build.
        context:
          type: object
          additionalProperties: true
          description: >-
            Optional. The `execution.context` object from the order build,
            passed unchanged. Validated against the signed transaction when
            present. For a Forecast swap it carries `jupiterSwapRequestId` and
            `ownerPubkey`.
        requestId:
          type: string
          description: Optional client correlation ID, echoed back in the response.
      required:
        - signedTransaction
    ExecuteResponse:
      type: object
      properties:
        status:
          type: string
          enum:
            - Success
            - Failed
          description: Transaction execution result.
        signature:
          type: string
          nullable: true
          description: Transaction signature (base58) when execution succeeds.
        error:
          type: string
          nullable: true
          description: Error message if the transaction failed, or `null` on success.
        requestId:
          type: string
          description: Echoed from the request when provided.
      required:
        - status
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