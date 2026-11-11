> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# List DCA Orders

> List your DCA orders with full event history



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml get /orders/history/dca
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
  /orders/history/dca:
    get:
      summary: List DCA orders
      description: List the authenticated wallet's DCA orders with full event history.
      parameters:
        - name: state
          in: query
          schema:
            type: string
            enum:
              - active
              - past
          description: >-
            active = depositing/active/executing/withdrawing; past =
            completed/cancelled/deposit_failed
        - name: mint
          in: query
          schema:
            type: string
          description: Filter by mint (matches inputMint or outputMint)
        - name: limit
          in: query
          schema:
            type: number
            default: 20
            minimum: 1
            maximum: 100
        - name: offset
          in: query
          schema:
            type: number
            default: 0
        - name: sort
          in: query
          schema:
            type: string
            enum:
              - updated_at
              - created_at
              - next_fill_at
            default: updated_at
        - name: dir
          in: query
          schema:
            type: string
            enum:
              - asc
              - desc
            default: desc
      responses:
        '200':
          description: Paginated DCA order history
          content:
            application/json:
              schema:
                type: object
                properties:
                  orders:
                    type: array
                    items:
                      $ref: '#/components/schemas/DcaHistoryItem'
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
              example:
                orders:
                  - id: 019f0530-36a7-77da-a38f-ed9414bb996b
                    requestId: b7d1b2a2-c8b3-4c21-a938-d1499260d218
                    userPubkey: 8gs8rXavMQRqoAoMhCfo79r2wndgSn93FLSWErdHp82n
                    vaultPubkey: 66xK9AE6oFpGAyn9eCWE3LtYwUZKkeUxbfjFSNmFnkBo
                    inputMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
                    outputMint: So11111111111111111111111111111111111111112
                    inputAmountInitial: '20000000'
                    inputAmountRemaining: '20000000'
                    amountPerRound: '10000000'
                    outputAmountTotal: '0'
                    inputAmountUsed: '0'
                    orderType: price_conditional
                    minPriceUsd: 50
                    maxPriceUsd: 1000
                    triggerMint: So11111111111111111111111111111111111111112
                    retryWindowSeconds: 7200
                    jlEnabled: false
                    jlYieldUsd: null
                    numberOfRounds: 2
                    intervalSeconds: 86400
                    beginFillAt: '2026-06-26T20:27:46.015Z'
                    nextFillAt: '2026-06-26T20:27:46.015Z'
                    lastFillAt: null
                    roundsFilled: 0
                    fillPercent: 0
                    state: active
                    displayState: active
                    createdAt: '2026-06-26T18:27:53.665Z'
                    updatedAt: '2026-06-26T18:27:54.529Z'
                    events:
                      - type: deposit
                        timestamp: '2026-06-26T18:27:54.529Z'
                        txSignature: >-
                          PPbp1QbbKDPjpHM3ScKyZBnCtLKKqmWwDeVj7USdBeEoNQsA1hkN35UQfrY2cQpyAB1pfVi9FdfciYdz7qUXBQA
                        inputAmount: '20000000'
                        state: success
                pagination:
                  total: 1
                  limit: 20
                  offset: 0
        '401':
          description: Unauthorized
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    DcaHistoryItem:
      type: object
      properties:
        id:
          type: string
          description: DCA order UUID
        requestId:
          type: string
        userPubkey:
          type: string
        vaultPubkey:
          type: string
        inputMint:
          type: string
        outputMint:
          type: string
        inputAmountInitial:
          type: string
          description: Total deposit amount in smallest units
        inputAmountRemaining:
          type: string
          description: Amount still held in the vault
        amountPerRound:
          type: string
          description: Amount swapped each round (the last round may include dust)
        outputAmountTotal:
          type: string
          description: Total output received across all successful rounds
        inputAmountUsed:
          type: string
          description: Total input consumed by successful rounds
        orderType:
          type: string
          enum:
            - time_based
            - price_conditional
        minPriceUsd:
          type: number
          nullable: true
        maxPriceUsd:
          type: number
          nullable: true
        triggerMint:
          type: string
          nullable: true
        retryWindowSeconds:
          type: number
          nullable: true
          description: >-
            Retry window per round in seconds, clamp(intervalSeconds/2, 30,
            7200)
        jlEnabled:
          type: boolean
          description: >-
            Whether Earn While You Wait is on for this order (idle funds earn
            Jupiter Lend yield between rounds).
        jlYieldUsd:
          type: number
          nullable: true
          description: >-
            Yield earned so far in USD for an Earn While You Wait order, from
            the ~15 min price cache. null when not JL-enabled or a required
            price is uncached.
        numberOfRounds:
          type: number
        intervalSeconds:
          type: number
        beginFillAt:
          type: string
          description: ISO-8601 time the first round is scheduled
        nextFillAt:
          type: string
          description: ISO-8601 time of the next scheduled round
        lastFillAt:
          type: string
          nullable: true
        roundsFilled:
          type: number
        fillPercent:
          type: number
          description: Progress from 0 to 1
        state:
          type: string
          enum:
            - depositing
            - deposit_failed
            - active
            - executing
            - withdrawing
            - completed
            - cancelled
        displayState:
          type: string
          enum:
            - pending
            - active
            - executing
            - pending_withdraw
            - completed
            - cancelled
            - failed
        createdAt:
          type: string
        updatedAt:
          type: string
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
              timestamp:
                type: string
              txSignature:
                type: string
              roundNumber:
                type: number
              inputAmount:
                type: string
              outputAmount:
                type: string
              state:
                type: string
                enum:
                  - success
                  - failed
                  - rescheduled
                  - pending
              attemptCount:
                type: number
      required:
        - id
        - userPubkey
        - vaultPubkey
        - inputMint
        - outputMint
        - orderType
        - state
        - displayState
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