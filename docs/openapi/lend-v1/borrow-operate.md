> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Operate

> Request a base64-encoded unsigned borrow operate transaction (deposit, borrow, repay, withdraw)



## OpenAPI

````yaml openapi-spec/lend/borrow.yaml post /borrow/operate
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
  /borrow/operate:
    post:
      tags:
        - Borrow
      summary: Build unsigned operate transaction
      description: >
        Builds a base64-encoded unsigned versioned transaction for a borrow
        operate action.


        Positive `colAmount` / `debtAmount` values deposit collateral or borrow
        debt.

        Negative values withdraw collateral or pay back debt.
      operationId: buildBorrowOperateTransaction
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
          description: Unsigned operate transaction
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OperateTransactionResponse'
              example:
                nftId: 9062
                transaction: >-
                  AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAgSLx7tPDvhw1CP7ZRClws8DN/LJyixatJZus2rQZ5pAG4cj71J70/rvu0J/yTO1674UGxqeljWsbxMzyBLGGTIhSXDKb3wVHYzXo0QFPryr6iEn/iNAYzX5PAlDQtaHW8lZfXf/XyEr3Wopfe9zqFiLis6v4f8U6hxuQhWGjqLG94CERPvP3H/BclenozhcxbxUXvL+eHEhCG764X4ME0EGHl8szytbAyAc7IdpMQ3ZAWF7Tn2sEZyyHpR1gWMtyb6oAuWSMFFuPlFmrZgneVXPFM1IT/BLHA9RbwtgiUIANaxKCIpkUG7pyNCcpG772SuCkkknsAKWSbiImP8gVvFqPeA4oganO3BaPj9y3TzLqfiwk9dDCxaIuRcmdJ+MiQSCv4el+RWy7kuZColtQumOW/ySBZLynRpY6hFkqzWzMMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEcn8RD5rZyCj3b2Rq49ZiVsmScRvvn3bU9qhBHACBPVR6mnvvwFsKRZEO8A58f97e1/yRiQ5ahJERwinC6lqslJomRsm5Gr9y9W+wdHAcnDKneVfnQTMTEQq+wdbs8IgIyXJY9OJInxuz0QKRSODYMLWhOZ2v8QhASOe9jb6fhZxvp6877brTo9ZfNqq8l0MbG75MLS9uDkfKYCA0UvXWEK/huRLkhelf0V6yk33/xLN6MW0KY4Ev8Cukm0xsGNHgbd9uHXZaGT2cvhRs7reawctIXtX1s3kTqM9YV+/wCpjWK6m/JZmhh7wDcE7nwe+sC7C7VoO3NCsHbuaiJo3IYBEBEABwIPCwEFBAgMBgMJDREOChDyI8aJUuHythAnAAAAAAAA
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
    OperateTransactionResponse:
      required:
        - nftId
        - transaction
      type: object
      properties:
        nftId:
          type: integer
          description: Position NFT id affected by the operation.
        transaction:
          type: string
          description: Base64-encoded unsigned versioned transaction.
    Market:
      type: string
      default: main
      enum:
        - main
        - ethena
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.