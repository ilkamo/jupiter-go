> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Log Out All

> Revoke every session for the wallet



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /auth/logout-all
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
  /auth/logout-all:
    post:
      summary: Revoke every session for the wallet
      description: >
        Invalidate every refresh token issued for the authenticated wallet,
        forcing all devices to

        re-authenticate. Requires the access token as a Bearer credential.
        Access tokens already issued

        remain valid until they expire (up to 15 minutes).
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
            example: {}
      responses:
        '200':
          description: All sessions revoked.
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                    enum:
                      - true
                required:
                  - success
              example:
                success: true
        '401':
          description: Missing or invalid access token
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
  schemas:
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        details:
          type: object
          additionalProperties:
            type: string
      required:
        - error
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