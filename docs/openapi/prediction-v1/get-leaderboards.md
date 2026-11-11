> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Leaderboards

> Get leaderboard rankings by various metrics



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /leaderboards
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
  /leaderboards:
    get:
      tags:
        - Leaderboards
      summary: Get leaderboards
      description: 'Ranked traders for a period and metric: P&L, volume or win rate.'
      parameters:
        - schema:
            type: string
            enum:
              - all_time
              - weekly
              - monthly
            default: all_time
          required: false
          description: 'Ranking window: `all_time`, `weekly` or `monthly`.'
          name: period
          in: query
        - schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 100
          required: false
          description: Maximum number of entries to return.
          name: limit
          in: query
        - schema:
            type: string
            enum:
              - pnl
              - volume
              - win_rate
            default: pnl
          required: false
          description: 'Ranking metric: `pnl`, `volume` or `win_rate`.'
          name: metric
          in: query
      responses:
        '200':
          description: Leaderboard entries
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        ownerPubkey:
                          type: string
                        realizedPnlUsd:
                          type: string
                        totalVolumeUsd:
                          type: string
                        predictionsCount:
                          type: number
                        correctPredictions:
                          type: number
                        wrongPredictions:
                          type: number
                        winRatePct:
                          type: string
                        period:
                          type: string
                        periodStart:
                          type: string
                          nullable: true
                        periodEnd:
                          type: string
                          nullable: true
                      required:
                        - ownerPubkey
                        - realizedPnlUsd
                        - totalVolumeUsd
                        - predictionsCount
                        - correctPredictions
                        - wrongPredictions
                        - winRatePct
                        - period
                        - periodStart
                        - periodEnd
                  summary:
                    type: object
                    properties:
                      all_time:
                        type: object
                        properties:
                          totalVolumeUsd:
                            type: string
                          predictionsCount:
                            type: number
                        required:
                          - totalVolumeUsd
                          - predictionsCount
                      weekly:
                        type: object
                        properties:
                          totalVolumeUsd:
                            type: string
                          predictionsCount:
                            type: number
                        required:
                          - totalVolumeUsd
                          - predictionsCount
                      monthly:
                        type: object
                        properties:
                          totalVolumeUsd:
                            type: string
                          predictionsCount:
                            type: number
                        required:
                          - totalVolumeUsd
                          - predictionsCount
                    required:
                      - all_time
                      - weekly
                      - monthly
                required:
                  - data
                  - summary
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.