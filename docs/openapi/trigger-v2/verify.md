> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Verify Challenge

> Submit the signed challenge to receive a JWT token



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /auth/verify
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
  /auth/verify:
    post:
      summary: Verify signed challenge and receive tokens
      description: >
        Submit the signed challenge to receive tokens. Set `authMode` to select
        the token model:


        - `access_refresh` (recommended): returns a 15-minute `accessToken` plus
        a rotating `refreshToken`.

        - `access_only` (deprecated, default when `authMode` is omitted):
        returns a single `token` valid for
          24 hours, with `Deprecation` and `Link` response headers. This mode will be removed; migrate to `access_refresh`.

        Include the access token in authenticated requests via the
        `Authorization: Bearer` header.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              oneOf:
                - type: object
                  title: MessageVerification
                  properties:
                    type:
                      type: string
                      enum:
                        - message
                    walletPubkey:
                      type: string
                    signature:
                      type: string
                      description: Base58-encoded signature
                    authMode:
                      type: string
                      enum:
                        - access_only
                        - access_refresh
                      default: access_only
                      description: >-
                        Token model. Defaults to the deprecated access_only when
                        omitted. Send access_refresh for refresh tokens.
                  required:
                    - type
                    - walletPubkey
                    - signature
                - type: object
                  title: TransactionVerification
                  properties:
                    type:
                      type: string
                      enum:
                        - transaction
                    walletPubkey:
                      type: string
                    signedTransaction:
                      type: string
                      description: Base64-encoded signed transaction
                    authMode:
                      type: string
                      enum:
                        - access_only
                        - access_refresh
                      default: access_only
                      description: >-
                        Token model. Defaults to the deprecated access_only when
                        omitted. Send access_refresh for refresh tokens.
                  required:
                    - type
                    - walletPubkey
                    - signedTransaction
            example:
              type: message
              walletPubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              signature: 5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d...
              authMode: access_refresh
      responses:
        '200':
          description: >
            Tokens for the requested auth mode. The access_only variant is
            returned with

            `Deprecation` (an RFC 8594 HTTP-date, e.g. `Tue, 21 Jul 2026
            00:00:00 GMT`) and

            `Link` (`rel="successor-version"` naming the absolute refresh URL

            `https://api.jup.ag/trigger/v2/auth/refresh`) response headers.
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/AccessRefreshTokens'
                    title: AccessRefreshResult
                  - title: AccessOnlyResult
                    type: object
                    description: >-
                      Deprecated. Returned when authMode is omitted or set to
                      access_only.
                    properties:
                      authMode:
                        type: string
                        enum:
                          - access_only
                      token:
                        type: string
                        description: JWT token (24h TTL). Deprecated; use access_refresh.
                    required:
                      - authMode
                      - token
              example:
                authMode: access_refresh
                accessToken: eyJhbGciOiJIUzI1NiIs...
                refreshToken: 01a08512-95d8-70ac-bab5-20c73aa002bb.6lr8kn3eMIsNGR9rXR1...
                expiresAt: '2026-09-16T14:12:43.000Z'
                tokenType: Bearer
        '400':
          description: Invalid request body or expired challenge
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Invalid signature
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