> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# DBC Pool Addresses by Mint

> Request for pool addresses for a given token mint



## OpenAPI

````yaml openapi-spec/studio/studio.yaml get /dbc-pool/addresses/{mint}
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
  /dbc-pool/addresses/{mint}:
    get:
      summary: dbc-pool-addresses-by-mint
      description: |
        Request for pool addresses for a given token mint
      parameters:
        - name: mint
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      dammv2PoolAddress:
                        type: string
                        nullable: true
                      dbcPoolAddress:
                        type: string
                        nullable: true
                      configKey:
                        type: string
                        nullable: true
              example:
                success: true
                data:
                  dammv2PoolAddress: PoolAddress11111111111111111111111111111111
                  dbcPoolAddress: DBCPoolAddress111111111111111111111111111111
                  configKey: ConfigKeyAddressExample11111111111111111111
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  errors:
                    type: object
                    description: Validation errors
        '404':
          description: Pool addresses not found
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    example: Pool addresses not found
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
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