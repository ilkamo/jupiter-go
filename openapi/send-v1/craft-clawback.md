> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Craft Clawback

> Request for a base64-encoded unsigned Send transaction



## OpenAPI

````yaml openapi-spec/send/send.yaml post /craft-clawback
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
  /craft-clawback:
    post:
      summary: craft-clawback
      description: |
        Request for a base64-encoded unsigned Send transaction
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - invitePDA
                - sender
              properties:
                invitePDA:
                  type: string
                  description: >
                    - Program Derived Address derived by `"invite"` and the
                    public key of the invite

                    - Do note that, the invitePDA requires to be generated
                    client side, [refer to docs](/docs/send)

                    - Please handle invite code and secret key with highest
                    security as user funds are involved
                sender:
                  type: string
                  description: >
                    - Public key address of the sender

                    - Since this address funded the invite, it will be the
                    receiving address of the full amount
            example:
              invitePDA: 8mF4HKoPQNmtbRvTRqotb4LW3pz5rR3eB7owZL9FVKIW
              sender: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                required:
                  - tx
                properties:
                  tx:
                    type: string
                    description: |
                      - Base64-encoded unsigned Send transaction
              example:
                tx: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAEN...
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