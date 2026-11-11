> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Claim Position

> Request an unsigned transaction to claim payout from a winning position



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml post /positions/{positionPubkey}/claim
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
  /positions/{positionPubkey}/claim:
    post:
      tags:
        - Positions
      summary: Claim position
      description: >-
        Build an unsigned transaction that claims the payout of a winning
        position after the market has resolved. Each winning contract pays out
        $1. Sign and send the returned transaction.
      parameters:
        - schema:
            type: string
            minLength: 32
            description: Position account public key
          required: true
          description: Position account public key
          name: positionPubkey
          in: path
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ClaimPositionRequest'
      responses:
        '200':
          description: Unsigned transaction to claim payout for a winning position
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ClaimPositionResponse'
        '400':
          description: Unable to build claim payout transaction
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Position not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  schemas:
    ClaimPositionRequest:
      type: object
      properties:
        ownerPubkey:
          type: string
          minLength: 32
          description: Position owner public key
    ClaimPositionResponse:
      type: object
      properties:
        transaction:
          type: string
          description: Base64 encoded transaction
        txMeta:
          type: object
          properties:
            blockhash:
              type: string
            lastValidBlockHeight:
              type: integer
              minimum: 0
          required:
            - blockhash
            - lastValidBlockHeight
        position:
          type: object
          properties:
            positionPubkey:
              type: string
              description: Position account public key
            marketPubkey:
              type: string
              description: Market account public key
            userPubkey:
              type: string
              description: User public key requesting claim
            ownerPubkey:
              type: string
              description: Position owner public key
            isYes:
              type: boolean
            contracts:
              type: string
              description: >-
                Legacy whole-contract quantity settled. Use contractsMicro or
                contractsDecimal for the exact fractional quantity.
            contractsMicro:
              type: string
              description: >-
                Contracts settled in micro-contract units (u64 as string,
                1000000 = 1 contract)
            contractsDecimal:
              type: string
              description: Contracts settled as a decimal string
            payoutAmountUsd:
              type: string
              description: Payout amount in micro USD (u64 as string)
          required:
            - positionPubkey
            - marketPubkey
            - userPubkey
            - ownerPubkey
            - isYes
            - contracts
            - payoutAmountUsd
      required:
        - transaction
        - txMeta
        - position
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