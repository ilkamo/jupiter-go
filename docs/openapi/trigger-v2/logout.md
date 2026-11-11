> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Log Out

> Revoke the refresh token for one session



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /auth/logout
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
  /auth/logout:
    post:
      summary: Revoke one session
      description: >
        Revoke the refresh token for a single session. No authentication is
        required and the endpoint always

        returns `200`. Access tokens already issued remain valid until they
        expire (up to 15 minutes).
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                refreshToken:
                  type: string
                  description: The refresh token to revoke.
              required:
                - refreshToken
            example:
              refreshToken: 01a08512-95d8-70ac-bab5-20c73aa002bb.UtUzQa8kUgtu2XR2...
      responses:
        '200':
          description: Session revoked.
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