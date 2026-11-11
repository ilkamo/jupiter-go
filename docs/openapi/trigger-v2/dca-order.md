> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get DCA Order

> Get a single DCA order by ID, including fills and events



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml get /orders/history/dca/{id}
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
  /orders/history/dca/{id}:
    get:
      summary: Get DCA order
      description: >
        Get a single DCA order by ID, including fill history and events. The
        path requires

        the `dca` segment.
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: DCA order UUID
      responses:
        '200':
          description: DCA order details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DcaHistoryItem'
              example:
                id: 019f053d-3f0c-70ad-bb5e-76d633291f80
                requestId: 1cc1ad82-866e-492d-a092-18859565945d
                userPubkey: 8gs8rXavMQRqoAoMhCfo79r2wndgSn93FLSWErdHp82n
                vaultPubkey: 66xK9AE6oFpGAyn9eCWE3LtYwUZKkeUxbfjFSNmFnkBo
                inputMint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
                outputMint: So11111111111111111111111111111111111111112
                inputAmountInitial: '20000000'
                inputAmountRemaining: '10000000'
                amountPerRound: '10000000'
                outputAmountTotal: '137589159'
                inputAmountUsed: '10000000'
                orderType: time_based
                minPriceUsd: null
                maxPriceUsd: null
                triggerMint: null
                retryWindowSeconds: 1800
                jlEnabled: false
                jlYieldUsd: null
                numberOfRounds: 2
                intervalSeconds: 3600
                beginFillAt: '2026-06-26T18:42:07.750Z'
                nextFillAt: '2026-06-26T19:44:57.469Z'
                lastFillAt: '2026-06-26T18:42:11.469Z'
                roundsFilled: 1
                fillPercent: 0.5
                state: active
                displayState: active
                createdAt: '2026-06-26T18:42:07.782Z'
                updatedAt: '2026-06-26T18:42:11.469Z'
                events:
                  - type: fill
                    timestamp: '2026-06-26T18:42:11.469Z'
                    txSignature: >-
                      HuNMpjFkTjVrCYMN6gTPWV9ay3GSxTxM7sZj4WNzjvzQoe4J5CR71MnTpuMKH9rtkDoDCTWu6rYRM12oR9cpbb4
                    roundNumber: 1
                    inputAmount: '10000000'
                    outputAmount: '137589159'
                    state: success
                  - type: deposit
                    timestamp: '2026-06-26T18:42:08.880Z'
                    txSignature: >-
                      32hAuVdZ57y5TUEx4vdV75zJyTqoCVqAby4eiayrmUW1TCuwoAvPRBcnhPMLN1YpfCp4Rc6HiLXKsi3NF5iWVcrn
                    inputAmount: '20000000'
                    state: success
        '401':
          description: Unauthorized
        '404':
          description: Order not found or not owned by the authenticated wallet
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