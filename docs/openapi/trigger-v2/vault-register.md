> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Register Vault

> Register a new trigger vault for the account



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml get /vault/register
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
  /vault/register:
    get:
      summary: Register a new vault
      description: |
        Create a new Privy-managed vault wallet for your account.
        Call this once per wallet. Subsequent calls return the existing vault.
      responses:
        '201':
          description: Vault created
          content:
            application/json:
              schema:
                type: object
                properties:
                  userPubkey:
                    type: string
                  vaultPubkey:
                    type: string
                  privyVaultId:
                    type: string
                required:
                  - userPubkey
                  - vaultPubkey
                  - privyVaultId
        '409':
          description: Vault already exists
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