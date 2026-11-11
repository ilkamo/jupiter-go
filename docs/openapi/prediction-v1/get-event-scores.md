> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Live Scores

> Get the latest live scores for a set of events



## OpenAPI

````yaml openapi-spec/prediction/prediction.yaml get /events/scores
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
  /events/scores:
    get:
      tags:
        - Events
      summary: Get live scores
      description: >-
        Latest live scores for several sports events in one call. Pass the event
        ids as a comma-separated `eventIds` list; events without a score are
        omitted from the response.
      parameters:
        - schema:
            type: string
            minLength: 1
            example: POLY-351731,POLY-30615
          required: true
          description: Comma-separated event IDs to fetch live scores for (max 100)
          name: eventIds
          in: query
      responses:
        '200':
          description: >-
            Latest live scores for the given event IDs. Events without a score
            row are omitted.
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/GameScore'
                required:
                  - data
components:
  schemas:
    GameScore:
      type: object
      properties:
        eventId:
          type: string
          description: Application-level event identifier
        gameId:
          type: string
        leagueAbbreviation:
          type: string
          nullable: true
        homeTeam:
          type: string
          nullable: true
        awayTeam:
          type: string
          nullable: true
        score:
          type: string
          nullable: true
          description: Raw score string from the provider (e.g. "2-0")
        period:
          type: string
          nullable: true
          description: Current period (e.g. "2H")
        elapsed:
          type: string
          nullable: true
          description: Elapsed time within the game (e.g. "90")
        status:
          type: string
          nullable: true
          description: Provider status string (e.g. "InProgress")
        live:
          type: boolean
          description: True while the game is in progress
        ended:
          type: boolean
          description: True once the game has finished
        finishedTimestamp:
          type: string
          nullable: true
          description: ISO 8601 timestamp when the game finished, if ended
        updatedAt:
          type: string
          description: ISO 8601 timestamp of the last score update
      required:
        - eventId
        - gameId
        - live
        - ended
        - updatedAt
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.