> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Redeem Instructions

> Request for the instruction of an earn redeem transaction to redeem shares



## OpenAPI

````yaml openapi-spec/lend/lend.yaml post /earn/redeem-instructions
openapi: 3.0.3
info:
  title: Jupiter Lend API
  version: 1.0.0
  description: Jupiter Lend API Schema
servers:
  - url: https://api.jup.ag/lend/v1
    description: Jupiter Lend API endpoint
security: []
paths:
  /earn/redeem-instructions:
    post:
      summary: redeem-instructions
      description: >
        Request for the instruction of an earn redeem transaction to redeem
        shares
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/EarnSharesRequestBody'
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/InstructionResponse'
              example:
                programId: jup3YeL8QhtSx1e253b2FDvsMNC87fDrgQZivbrndc9
                accounts:
                  - pubkey: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
                    isSigner: true
                    isWritable: true
                data: 6Ge5AAAAAAA...
      security:
        - ApiKeyAuth: []
components:
  schemas:
    EarnSharesRequestBody:
      type: object
      required:
        - asset
        - signer
        - shares
      properties:
        asset:
          type: string
          default: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer:
          type: string
          default: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        shares:
          type: string
          default: '1000000'
      example:
        asset: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
        signer: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        shares: '1000000'
    InstructionResponse:
      type: object
      required:
        - programId
        - accounts
        - data
      properties:
        programId:
          type: string
        accounts:
          type: array
          items:
            $ref: '#/components/schemas/AccountMeta'
        data:
          type: string
    AccountMeta:
      type: object
      required:
        - pubkey
        - isSigner
        - isWritable
      properties:
        pubkey:
          type: string
        isSigner:
          type: boolean
        isWritable:
          type: boolean
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.