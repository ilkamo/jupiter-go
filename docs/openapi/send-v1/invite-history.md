> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Invite History

> Request for the invite history of an address



## OpenAPI

````yaml openapi-spec/send/send.yaml get /invite-history
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
  /invite-history:
    get:
      summary: invite-history
      description: |
        Request for the invite history of an address
      parameters:
        - name: address
          in: query
          description: |
            - Pubkey can be sender or recipient
          required: true
          schema:
            type: string
        - name: page
          in: query
          description: |
            - Pagination of response
          schema:
            type: integer
            minimum: 1
            default: 1
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/InviteDataResponse'
              example:
                invites:
                  - invite_signer: 7nE9GJoYHNmtaQvTQpota3KV2oz4pQ2dA6nvYK8EUJHV
                    invite_pda: 8mF4HKoPQNmtbRvTRqotb4LW3pz5rR3eB7owZL9FVKIW
                    sender: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
                    expiry: '2025-03-22T10:30:00.000Z'
                    amount: '1000000'
                    creation_tx: 4xKpN2...
                    deletion_tx: 5yLqO3...
                    receiver: 9pG5ILqRSOnucCwUSRspuc5MV4pR4sD8fC8qxNL0GWJX
                    action: claimed
                    creation_time: '2025-03-15T10:30:00.000Z'
                    deletion_time: '2025-03-16T08:15:00.000Z'
                    mint: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
                    confirmed: true
                hasMoreData: false
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
  schemas:
    InviteDataResponse:
      type: object
      required:
        - invites
        - hasMoreData
      properties:
        invites:
          type: array
          items:
            type: object
            required:
              - invite_signer
              - invite_pda
              - sender
              - expiry
              - amount
              - creation_tx
              - deletion_tx
              - receiver
              - action
              - creation_time
              - deletion_time
              - mint
              - confirmed
            properties:
              invite_signer:
                type: string
              invite_pda:
                type: string
              sender:
                type: string
              expiry:
                type: string
              amount:
                type: string
              creation_tx:
                type: string
              deletion_tx:
                type: string
                nullable: true
              receiver:
                type: string
                nullable: true
              action:
                type: string
                nullable: true
              creation_time:
                type: string
              deletion_time:
                type: string
                nullable: true
              mint:
                type: string
                nullable: true
              confirmed:
                type: integer
                nullable: true
        hasMoreData:
          type: boolean
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.