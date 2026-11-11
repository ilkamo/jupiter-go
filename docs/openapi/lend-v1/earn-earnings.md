> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Earnings

> Request for the earnings of one or multiple positions of a user



## OpenAPI

````yaml openapi-spec/lend/lend.yaml get /earn/earnings
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
  /earn/earnings:
    get:
      summary: earnings
      description: |
        Request for the earnings of one or multiple positions of a user
      parameters:
        - name: user
          in: query
          description: |
            User wallet address
          required: true
          schema:
            type: string
          example: HYbxGkNvEwvZ14RzJHPB9h3dWfXjxwAEhkyzJRHx1hBf
        - name: positions
          in: query
          description: |
            User token positions (comma-separated)
          required: true
          schema:
            type: string
          example: >-
            9BEcn9aPEmhSPbPQeFGjidRiEKki46fVQDyPpSQXPA2D,2uQsyo1fXXQkDtcpXnLofWy88PxcvnfH2L8FPSE62FVU
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserEarningsResponse'
              example:
                - address: So11111111111111111111111111111111111111112
                  ownerAddress: BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV
                  earnings: 34483000
                  slot: 345678901
      security:
        - ApiKeyAuth: []
components:
  schemas:
    UserEarningsResponse:
      type: object
      required:
        - address
        - ownerAddress
        - earnings
        - slot
      properties:
        address:
          type: string
        ownerAddress:
          type: string
        earnings:
          type: number
        slot:
          type: number
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.