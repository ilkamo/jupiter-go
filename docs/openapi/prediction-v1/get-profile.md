> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Profile

> Get profile statistics for a user



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /profiles/{ownerPubkey}
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
  /profiles/{ownerPubkey}:
    get:
      tags:
        - Profile
      summary: Get profile
      description: >-
        Trading statistics of a wallet: realised P&L, volume, prediction counts,
        active contracts and portfolio value. USD values are in micro USD.
      parameters:
        - schema:
            type: string
          required: true
          description: Wallet public key.
          name: ownerPubkey
          in: path
      responses:
        '200':
          description: Profile stats
          content:
            application/json:
              schema:
                type: object
                properties:
                  ownerPubkey:
                    type: string
                  realizedPnlUsd:
                    type: string
                  totalVolumeUsd:
                    type: string
                  predictionsCount:
                    type: string
                  correctPredictions:
                    type: string
                  wrongPredictions:
                    type: string
                  totalActiveContracts:
                    type: string
                    description: >-
                      Legacy whole-contract count. Use totalActiveContractsMicro
                      or totalActiveContractsDecimal for the exact value.
                  totalActiveContractsMicro:
                    type: string
                    description: >-
                      Active contracts in micro-contract units (1000000 = 1
                      contract)
                  totalActiveContractsDecimal:
                    type: string
                    description: Active contracts as a decimal string
                  totalPositionsValueUsd:
                    type: string
                required:
                  - ownerPubkey
                  - realizedPnlUsd
                  - totalVolumeUsd
                  - predictionsCount
                  - correctPredictions
                  - wrongPredictions
                  - totalActiveContracts
                  - totalPositionsValueUsd
        '404':
          description: Profile not found
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.