> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Operate Instructions

> Request the instructions and address lookup tables for a borrow operate transaction



## OpenAPI

````yaml openapi-spec/lend/borrow.yaml post /borrow/operate-instructions
openapi: 3.0.3
info:
  title: Jupiter Lend Borrow API
  description: >
    OpenAPI specification for Jupiter Lend borrow integrator endpoints.


    **Base URL:** `https://api.jup.ag/lend/v1`


    Share this file with documentation tools (ReadMe, Stoplight, Postman, Redoc,
    Mintlify,

    Swagger UI, etc.) to generate API reference docs.


    | Method | Path | Description |

    | ------ | ---- | ----------- |

    | GET | `/borrow/vaults` | List borrow vaults |

    | GET | `/borrow/positions` | List positions for wallet addresses |

    | POST | `/borrow/operate` | Build unsigned operate transaction |

    | POST | `/borrow/operate-instructions` | Build operate Solana instructions
    |


    **Notes**

    - Pass multiple wallets to `GET /borrow/positions` as a comma-separated
    `users` query value.

    - `colAmount` / `debtAmount` are signed integer strings in token base units.

    - Use `positionId: 0` in operate requests to open a new position.

    - Optional `market` query/body context: `main` (default) or `ethena`.
  version: 1.0.0
servers:
  - url: https://api.jup.ag/lend/v1
    description: Jupiter Lend API endpoint
security: []
tags:
  - name: Borrow
    description: Vault discovery, position lookup, and operate transaction building
paths:
  /borrow/operate-instructions:
    post:
      tags:
        - Borrow
      summary: Build operate instructions
      description: >-
        Returns Solana instructions and lookup table addresses for a borrow
        operate action.
      operationId: buildBorrowOperateInstructions
      parameters:
        - $ref: '#/components/parameters/market'
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/OperatePayload'
            example:
              vaultId: 1
              positionId: 0
              signer: 4AwWnVxWKdZCZ8zVGdD7S3EwzRHkVRiXJxQKzwjbBayj
              colAmount: '10000'
              debtAmount: '100'
        required: true
      responses:
        '200':
          description: Operate instructions
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OperateInstructionsResponse'
              example:
                nftId: 9062
                instructions:
                  - programId: jupr81YtYssSyPt8jbnGuiWon5f6x9TcDEFxYe3Bdzi
                    accounts:
                      - pubkey: GpLjUnj5q8V6jkaaT7q26rQ4hzn7KYoebLRWfk16hTwh
                        isSigner: true
                        isWritable: true
                    data: xRQKAWGgsVsBAF8jAAA=
                addressLookupTableAddresses:
                  - 3eB86bwSJanGwLo5WoMQZL7Bk7uubCRggZKshJKJVKwx
      security:
        - ApiKeyAuth: []
components:
  parameters:
    market:
      name: market
      in: query
      description: Lend market identifier. Defaults to `main`.
      required: false
      style: form
      explode: true
      schema:
        $ref: '#/components/schemas/Market'
  schemas:
    OperatePayload:
      required:
        - colAmount
        - debtAmount
        - positionId
        - signer
        - vaultId
      type: object
      properties:
        vaultId:
          type: integer
          description: Target vault id.
          example: 1
        positionId:
          type: integer
          description: Existing position id. Use `0` when opening a new position.
          example: 0
        positionOwner:
          type: string
          description: Optional position owner when operating on behalf of another wallet.
          example: 4AwWnVxWKdZCZ8zVGdD7S3EwzRHkVRiXJxQKzwjbBayj
        signer:
          type: string
          description: Transaction fee payer and signer wallet address.
          example: 4AwWnVxWKdZCZ8zVGdD7S3EwzRHkVRiXJxQKzwjbBayj
        colAmount:
          pattern: ^-?[0-9]+$
          type: string
          description: >-
            Collateral delta in supply-token base units. Positive deposits
            collateral, negative withdraws. Use
            -170141183460469231731687303715884105728 (MIN_I128) to withdraw all
            collateral.
          example: '10000'
        debtAmount:
          pattern: ^-?[0-9]+$
          type: string
          description: >-
            Debt delta in borrow-token base units. Positive borrows, negative
            repays. Use -170141183460469231731687303715884105728 (MIN_I128) to
            repay all debt.
          example: '100'
      description: >-
        Request body validated by `operateValidator` in
        `app/validators/borrowing.ts`.
    OperateInstructionsResponse:
      required:
        - addressLookupTableAddresses
        - instructions
        - nftId
      type: object
      properties:
        nftId:
          type: integer
        instructions:
          type: array
          items:
            $ref: '#/components/schemas/SolanaInstruction'
        addressLookupTableAddresses:
          type: array
          items:
            type: string
    Market:
      type: string
      default: main
      enum:
        - main
        - ethena
    SolanaInstruction:
      required:
        - accounts
        - data
        - programId
      type: object
      properties:
        programId:
          type: string
        accounts:
          type: array
          items:
            $ref: '#/components/schemas/SolanaAccountMeta'
        data:
          type: string
          description: Base64-encoded instruction data.
    SolanaAccountMeta:
      required:
        - isSigner
        - isWritable
        - pubkey
      type: object
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