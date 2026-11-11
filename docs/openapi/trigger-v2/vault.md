> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Vault

> Get vault information and balance for an account



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml get /vault
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
  /vault:
    get:
      summary: Get vault info
      description: Retrieve the vault associated with your authenticated wallet.
      responses:
        '200':
          description: Vault details
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
                  privyUserId:
                    type: string
                    nullable: true
                required:
                  - userPubkey
                  - vaultPubkey
                  - privyVaultId
              example:
                userPubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
                vaultPubkey: 7nE9GJoYHNmtaQvTQpota3KV2oz4pQ2dA6nvYK8EUJHV
                privyVaultId: vault-uuid
        '404':
          description: No vault found. Use /vault/register to create one.
      security:
        - ApiKeyAuth: []
          BearerAuth: []
components:
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