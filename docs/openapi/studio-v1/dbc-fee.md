> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# DBC Fee

> Request for unclaimed creator trading fees of a Dynamic Bonding Curve pool



## OpenAPI

````yaml openapi-spec/studio/studio.yaml post /dbc/fee
openapi: 3.0.3
info:
  title: Jupiter Studio API
  version: 2.0.0
  description: Jupiter Studio API Schema
servers:
  - url: https://api.jup.ag/studio/v1
    description: Jupiter Studio API Endpoint
security: []
paths:
  /dbc/fee:
    post:
      summary: dbc-fee
      description: >
        Request for unclaimed creator trading fees of a Dynamic Bonding Curve
        pool
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - poolAddress
              properties:
                poolAddress:
                  type: string
                  description: |
                    - Dynamic Bonding Curve pool address
            example:
              poolAddress: PoolAddress11111111111111111111111111111111
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                required:
                  - unclaimed
                  - total
                properties:
                  unclaimed:
                    type: string
                    description: |
                      - Unclaimed creator quote fee amount
                  total:
                    type: string
                    description: >
                      - Total trading quote fee amount divided by 2 (creator
                      share)
              example:
                unclaimed: '5000000'
                total: '25000000'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  errors:
                    type: object
                    additionalProperties:
                      type: array
                      items:
                        type: string
        '500':
          description: Internal server error
      security:
        - ApiKeyAuth: []
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.