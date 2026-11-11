> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# DBC Pool Submit

> Execute the signed transaction, and optionally upload content and header image



## OpenAPI

````yaml openapi-spec/studio/studio.yaml post /dbc-pool/submit
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
  /dbc-pool/submit:
    post:
      summary: dbc-pool-submit
      description: >
        Execute the signed transaction, and optionally upload content and header
        image
      requestBody:
        content:
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/SubmitDBCTransactionRequestBody'
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
                      mint:
                        type: string
                      configKey:
                        type: string
              example:
                success: true
                data:
                  mint: TokenMintAddressExample111111111111111111111
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
                  error:
                    type: string
        '500':
          description: Internal server error
      security:
        - ApiKeyAuth: []
components:
  schemas:
    SubmitDBCTransactionRequestBody:
      type: object
      required:
        - transaction
        - owner
      properties:
        transaction:
          type: string
          description: |
            - Base64-encoded signed transaction
        owner:
          type: string
          description: |
            - Owner wallet public key
        content:
          type: string
          description: >
            - Optional content description displayed on Studio dedicated token
            page

            - This is stored off-chain for our frontend to display

            - This is NOT the on-chain token metadata

            - The on-chain token metadata is done by making a PUT request to the
            presigned URL
        headerImage:
          type: string
          format: binary
          description: >
            - Optional header image file displayed on Studio token page

            - This is stored off-chain for our frontend to display

            - This is NOT the on-chain token metadata

            - The on-chain token metadata is done by making a PUT request to the
            presigned URL
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.