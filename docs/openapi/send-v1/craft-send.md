> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Send

> Request for a base64-encoded unsigned Send transaction



## OpenAPI

````yaml openapi-spec/send/send.yaml post /craft-send
openapi: 3.0.3
info:
  title: Jupiter Send API
  version: 1.0.0
  description: Jupiter Send API Schema
servers:
  - url: https://api.jup.ag/send/v1
    description: Jupiter Send API Endpoint
security: []
paths:
  /craft-send:
    post:
      summary: craft-send
      description: |
        Request for a base64-encoded unsigned Send transaction
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - inviteSigner
                - sender
                - amount
              properties:
                inviteSigner:
                  type: string
                  description: >
                    - Public key address generated from the hash of the invite
                    code

                    - Do note that, the invite code requires to be generated
                    client side, [refer to docs](/docs/send)

                    - Please handle invite code and secret key with highest
                    security as user funds are involved
                sender:
                  type: string
                  description: >
                    - Public key address of the sender

                    - This address will be funding the amount to be sent and the
                    token accounts involved
                amount:
                  type: string
                  description: >
                    - Amount in atomic value (before decimals)

                    - E.g. to send 1 USDC is equivalent of `amount='1000000'`

                    - Do note that there will be additional SOL amount being
                    used in the transaction for transaction fees and token
                    accounts
                mint:
                  type: string
                  description: |
                    - Defaults to WSOL mint
                    - Only pass in this parameter with other mint when required
            example:
              inviteSigner: 7nE9GJoYHNmtaQvTQpota3KV2oz4pQ2dA6nvYK8EUJHV
              sender: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
              amount: '1000000'
              mint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                required:
                  - tx
                  - expiry
                  - totalFeeLamports
                properties:
                  tx:
                    type: string
                    description: |
                      - Base64-encoded unsigned Send transaction
                  expiry:
                    type: string
                    description: |
                      - Unix timestamp of when the invite will expire
                      - Expired invites will automatically clawback to sender
                  totalFeeLamports:
                    type: string
                    description: >
                      - Amount of fees required for network's transaction fee
                      and token accounts
              example:
                tx: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
                expiry: '2025-03-22T10:30:00.000Z'
                totalFeeLamports: '5000'
        '400':
          description: Bad request body
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
        '500':
          description: Internal error
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
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