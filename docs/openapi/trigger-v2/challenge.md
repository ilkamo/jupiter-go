> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Request Challenge

> Request a sign-in challenge for wallet authentication



## OpenAPI

````yaml openapi-spec/trigger/v2/trigger.yaml post /auth/challenge
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
  /auth/challenge:
    post:
      summary: Request authentication challenge
      description: >
        Request a challenge to sign with your wallet. Use `message` for standard
        wallets

        or `transaction` for hardware wallets that only support transaction
        signing.

        Challenges expire after 5 minutes.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                walletPubkey:
                  type: string
                  description: Your Solana wallet public key
                type:
                  type: string
                  enum:
                    - message
                    - transaction
                  description: Challenge type. Use `transaction` for hardware wallets.
              required:
                - walletPubkey
                - type
            example:
              walletPubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              type: message
      responses:
        '200':
          description: Challenge for wallet signing
          content:
            application/json:
              schema:
                oneOf:
                  - type: object
                    properties:
                      type:
                        type: string
                        enum:
                          - message
                      challenge:
                        type: string
                        description: Message to sign
                    required:
                      - type
                      - challenge
                  - type: object
                    properties:
                      type:
                        type: string
                        enum:
                          - transaction
                      transaction:
                        type: string
                        description: Base64-encoded transaction with memo instruction
                    required:
                      - type
                      - transaction
              example:
                type: message
                challenge: >-
                  Sign this message to authenticate with Jupiter that you are
                  the owner of BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV.


                  This challenge expires at 2026-09-16T14:02:39.204Z with the
                  nonce of 01a0aa82-4504-76ed-8c94-4fb787c6a2a9.
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
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

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.