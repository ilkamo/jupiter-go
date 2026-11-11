> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Refresh Token

> Rotate the refresh token and get a new access token



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /auth/refresh
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
  /auth/refresh:
    post:
      summary: Rotate the refresh token and get a new access token
      description: >
        Exchange the current refresh token for a new access token and a new
        refresh token. Every call

        rotates the refresh token: replace both stored tokens with the response
        and discard the old ones.

        Repeating the same request within a short grace window (a network retry)
        is tolerated. Presenting an

        old refresh token outside that window, or after logout, returns `401`
        and revokes the session, at which

        point you must re-authenticate through the challenge flow.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                refreshToken:
                  type: string
                  description: The current refresh token.
              required:
                - refreshToken
            example:
              refreshToken: 01a08512-95d8-70ac-bab5-20c73aa002bb.6lr8kn3eMIsNGR9rXR1...
      responses:
        '200':
          description: New access token and rotated refresh token.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AccessRefreshTokens'
              example:
                authMode: access_refresh
                accessToken: eyJhbGciOiJIUzI1NiIs...
                refreshToken: 01a08512-95d8-70ac-bab5-20c73aa002bb.UtUzQa8kUgtu2XR2...
                expiresAt: '2026-09-16T14:12:47.000Z'
                tokenType: Bearer
        '400':
          description: Missing or malformed refresh token
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: >-
            Refresh token expired, revoked, or reused. Re-authenticate via
            /auth/challenge.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
components:
  schemas:
    AccessRefreshTokens:
      type: object
      description: >-
        Access token plus rotating refresh token, returned identically by
        access_refresh verify and by refresh.
      properties:
        authMode:
          type: string
          enum:
            - access_refresh
        accessToken:
          type: string
          description: >-
            JWT to send in the Authorization Bearer header. Expires 15 minutes
            after issue.
        refreshToken:
          type: string
          description: >-
            Opaque token used to obtain a new access token. Rotates on every
            /auth/refresh call.
        expiresAt:
          type: string
          format: date-time
          description: ISO 8601 timestamp when accessToken expires.
        tokenType:
          type: string
          enum:
            - Bearer
      required:
        - authMode
        - accessToken
        - refreshToken
        - expiresAt
        - tokenType
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

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.