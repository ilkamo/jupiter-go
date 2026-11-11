> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Order (Swap API)

> Get a swap quote and assembled transaction



## OpenAPI

````yaml openapi-spec/swap/v2/swap.yaml get /order
openapi: 3.0.3
info:
  title: Swap API V2
  version: 2.0.0
  description: >
    Jupiter Swap API V2. Three endpoints: `/order` (default happy path with
    managed execution),

    `/build` (advanced path returning raw swap instructions), and `/execute`
    (managed transaction execution for `/order`).
servers:
  - url: https://api.jup.ag/swap/v2
    description: Swap API V2 Endpoint
security: []
paths:
  /order:
    get:
      tags:
        - Swap
      summary: Get Order
      description: >
        Get a swap quote and assembled transaction. This is the default happy
        path with all Ultra benefits

        (predictive execution, RTSE, Jupiter Beam, gasless, multi-router
        competition).


        Without optional parameters, all routers compete for the best price
        (Metis, JupiterZ RFQ, Dflow, OKX).

        Setting `payer` to a wallet different from `taker` disables Dflow and
        OKX. Metis and JupiterZ RFQ still compete.
      operationId: GetOrder
      parameters:
        - $ref: '#/components/parameters/InputMintParameter'
        - $ref: '#/components/parameters/OutputMintParameter'
        - $ref: '#/components/parameters/AmountParameter'
        - name: taker
          in: query
          schema:
            type: string
          description: >
            - The public key of the wallet that will sign the transaction

            - If not provided, the response will contain a quote but no
            transaction

            - Must be present if you intend to sign and execute the transaction
            via `/execute`
          example: GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ
        - name: receiver
          in: query
          schema:
            type: string
          description: >
            - The public key of the account that will receive the output tokens

            - Must differ from `taker`

            - Expects a wallet address, not a token account

            - For non-SOL output: tokens are sent to the receiver's associated
            token account (ATA). If the ATA is not initialised, a create ATA
            instruction is added to the transaction

            - For SOL output: native SOL is transferred directly to the receiver

            - Does not support destination wSOL token accounts

            - Disables Dflow and OKX routing. Metis and JupiterZ RFQ still
            compete
        - name: swapMode
          in: query
          schema:
            type: string
            enum:
              - ExactIn
          description: |
            - Swap mode. Currently only `ExactIn` is supported.
        - name: slippageBps
          in: query
          schema:
            type: integer
            minimum: 0
            maximum: 10000
          description: >
            - Slippage tolerance in basis points (0-10000)

            - If not set, Jupiter automatically determines an appropriate
            slippage
          example: 50
        - name: referralAccount
          in: query
          schema:
            type: string
          description: |
            - Address of your referral account for the Jupiter referral project
            - Must be used together with `referralFee`
            - See the [Referral Program](/tool-kits/referral-program) for setup
        - name: referralFee
          in: query
          schema:
            type: number
            minimum: 50
            maximum: 255
          description: |
            - Referral fee in basis points (50-255)
            - Must be used together with `referralAccount`
          example: 50
        - name: payer
          in: query
          schema:
            type: string
          description: >
            - The public key of an account that will cover gas-related fees
            (signature fees, priority fees, and rent) on behalf of the taker

            - Has no effect when set equal to `taker`. Must be a different
            wallet to activate integrator-sponsored gas

            - Disables Dflow and OKX routing. Metis and JupiterZ RFQ still
            compete

            - On JupiterZ routes the market maker pays signature and priority
            fees, and `payer` only pays rent for the output token accounts the
            transaction creates. When the output is native SOL there are none,
            so `payer` is not a signer

            - Sign with `payer` only if it is a required signer of the returned
            transaction
          example: gasTzr94Pmp4Gf8vknQnqxeYxdgwFjbgdJa4msYRpnB
        - name: priorityFeeLamports
          in: query
          schema:
            type: number
          description: >
            - Priority fee in lamports

            - If not set, Jupiter automatically determines an appropriate
            priority fee

            - Setting this overrides the automatic optimisation
        - name: jitoTipLamports
          in: query
          schema:
            type: number
          description: |
            - Jito MEV tip in lamports for faster block inclusion
        - name: broadcastFeeType
          in: query
          schema:
            type: string
            enum:
              - maxCap
              - exactFee
          description: >
            - Fee cap strategy: `maxCap` treats the fee as a maximum, `exactFee`
            uses the exact amount

            - Ignored if neither `priorityFeeLamports` nor `jitoTipLamports` are
            set
        - name: excludeRouters
          in: query
          schema:
            type: string
          description: |
            - Comma-separated list of routers to exclude
            - Available routers: `metis`, `jupiterz`, `dflow`, `okx`
          example: jupiterz,dflow
        - name: excludeDexes
          in: query
          schema:
            type: string
          description: >
            - Comma-separated list of DEXes to exclude from the Metis router

            - **Important**: This only affects the Metis router, not other
            routers

            - For example: `excludeDexes=Raydium,Orca+V2,Meteora+DLMM`

            - Labels are case sensitive (`SolFi`, not `solfi`). Full list of
            DEXes here using the `/program-id-to-label` endpoint.

            - On `/order` only `excludeDexes` is honoured, not `dexes`. Use
            `/build` to restrict routing to specific DEXes.
        - name: maxSupportedTransactionVersion
          in: query
          schema:
            type: string
            enum:
              - '0'
              - '1'
            default: '0'
          description: >
            - The highest Solana transaction version your client can accept,
            `"0"` or `"1"`. Defaults to `"0"`.

            - This is a ceiling, not a force. You get v1 only when Metis wins
            the route and the order is not gasless; other routers and gasless
            orders (including automatic sponsorship for low-SOL takers) return
            v0. Wider v1 support is in progress.

            - Read the response `transactionVersion` to see the version you got.
            v1 transactions must be deserialized and signed with a v1-capable
            SDK (`@solana/kit` v8 or later).
          example: '1'
      responses:
        '200':
          description: Quote with optional assembled transaction
          content:
            application/json:
              schema:
                type: object
                properties:
                  mode:
                    type: string
                    description: '"ultra" or "manual" based on parameters used'
                  inputMint:
                    type: string
                  outputMint:
                    type: string
                  inAmount:
                    type: string
                  outAmount:
                    type: string
                  inUsdValue:
                    type: number
                  outUsdValue:
                    type: number
                  priceImpact:
                    type: number
                    description: >-
                      Price impact in percentage points (e.g. -0.1 = -0.1%).
                      Divide by 100 to convert to a decimal fraction.
                  swapUsdValue:
                    type: number
                  otherAmountThreshold:
                    type: string
                    description: Minimum output amount after slippage
                  swapMode:
                    type: string
                  slippageBps:
                    type: number
                  priceImpactPct:
                    type: string
                    description: >-
                      Deprecated: decimal price impact ratio. Use priceImpact
                      divided by 100 instead.
                    deprecated: true
                  routePlan:
                    type: array
                    items:
                      $ref: '#/components/schemas/RoutePlanStep'
                  referralAccount:
                    type: string
                  feeMint:
                    type: string
                  feeBps:
                    type: number
                    description: >-
                      Total fee rate charged for the swap in basis points. This
                      includes the Jupiter platform fee plus any additional
                      charges, such as gasless support cost recoup.
                  platformFee:
                    $ref: '#/components/schemas/PlatformFee'
                  signatureFeeLamports:
                    type: number
                  signatureFeePayer:
                    type: string
                    nullable: true
                  prioritizationFeeLamports:
                    type: number
                    description: Includes priority fees and tips (Jito, Nozomi)
                  prioritizationFeePayer:
                    type: string
                    nullable: true
                  rentFeeLamports:
                    type: number
                    description: Estimated rent fee
                  rentFeePayer:
                    type: string
                    nullable: true
                  swapType:
                    type: string
                    description: 'Deprecated: use router instead'
                    deprecated: true
                    enum:
                      - aggregator
                      - rfq
                      - aggregator+rfq
                      - dflow
                      - okx
                  router:
                    type: string
                    enum:
                      - metis
                      - jupiterz
                      - dflow
                      - okx
                    description: Which router won the quote
                  transactionVersion:
                    type: number
                    enum:
                      - 0
                      - 1
                    description: >
                      The Solana message version of `transaction`, `0` or `1`.
                      With `maxSupportedTransactionVersion=1` this is `1` only
                      when Metis wins the route and the order is not gasless,
                      and `0` otherwise (other routers, and gasless orders
                      including automatic sponsorship).
                  transaction:
                    type: string
                    nullable: true
                    description: >
                      - Base64-encoded transaction. Null if taker is not
                      provided.

                      - Empty string if taker is provided but transaction could
                      not be built (check errorCode).
                  lastValidBlockHeight:
                    type: string
                  gasless:
                    type: boolean
                    description: >
                      True when signature and priority fees are paid by a wallet
                      other than the taker. Set by three independent paths:

                      - Automatic Jupiter sponsorship: `signatureFeePayer =
                      gasTzr94Pmp4Gf8vknQnqxeYxdgwFjbgdJa4msYRpnB`

                      - JupiterZ market maker: `signatureFeePayer` is the MM
                      address (varies per quote); Jupiter's gas wallet can pay
                      the rent to create the taker's output token account when
                      referral fees are not set

                      - Integrator payer: `signatureFeePayer` is your `payer`
                      parameter on Metis routes. On JupiterZ routes
                      `signatureFeePayer` is the MM address and `rentFeePayer`
                      is your `payer`

                      For a deterministic opt-out, check `signatureFeePayer ==
                      taker` on every response.
                  requestId:
                    type: string
                    description: Unique request ID. Pass this to /execute.
                  totalTime:
                    type: number
                    description: Response time in milliseconds
                  taker:
                    type: string
                    nullable: true
                  quoteId:
                    type: string
                    description: Quote ID for RFQ swaps
                  maker:
                    type: string
                    description: Market maker address for RFQ swaps
                  expireAt:
                    type: string
                    description: Quote expiration timestamp for RFQ swaps
                  errorCode:
                    type: number
                    description: >
                      Present when `taker` is defined and `transaction` is the
                      empty string. The router quoted a price but could not
                      build a transaction. Match on `router` + `errorCode` to
                      identify the error.


                      Aggregator routers (router is metis, dflow, okx):

                      - 1: Insufficient funds

                      - 2: Insufficient SOL for gas

                      - 3: Swap below minimum for gasless


                      JupiterZ router (router is jupiterz):

                      - 1: Insufficient balance to fund the swap

                      - 2: Missing associated token account

                      - 3: Quote could not be built into a transaction
                  errorMessage:
                    type: string
                    description: >
                      Human-readable error description. Present when `taker` is
                      defined and `transaction` is the empty string.

                      Match on `router` + `errorCode` instead of this string, as
                      the message text may be parameterised.
                  error:
                    type: string
                    description: Duplicate of errorMessage for backwards compatibility
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
                  requestId:
                    type: string
                  error:
                    type: string
      security:
        - ApiKeyAuth: []
components:
  parameters:
    InputMintParameter:
      name: inputMint
      in: query
      required: true
      schema:
        type: string
      description: The mint address of the input token
      example: So11111111111111111111111111111111111111112
    OutputMintParameter:
      name: outputMint
      in: query
      required: true
      schema:
        type: string
      description: The mint address of the output token
      example: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
    AmountParameter:
      name: amount
      in: query
      required: true
      schema:
        type: string
      description: The amount to swap in the smallest unit of the input token
      example: '1000000000'
  schemas:
    RoutePlanStep:
      type: object
      properties:
        swapInfo:
          $ref: '#/components/schemas/SwapInfo'
        percent:
          type: number
          description: Percentage of total swap routed through this step
        bps:
          type: number
          description: Basis points of total swap routed through this step
        usdValue:
          type: number
          description: USD value of this step
      required:
        - swapInfo
        - percent
        - bps
    PlatformFee:
      type: object
      properties:
        amount:
          type: string
        feeBps:
          type: number
          description: >-
            Jupiter platform fee component in basis points. This is the fee
            Jupiter intends to make from the swap and does not include gasless
            support cost recoup.
        feeMint:
          type: string
    SwapInfo:
      type: object
      properties:
        ammKey:
          type: string
        label:
          type: string
        inputMint:
          type: string
        outputMint:
          type: string
        inAmount:
          type: string
        outAmount:
          type: string
      required:
        - ammKey
        - label
        - inputMint
        - outputMint
        - inAmount
        - outAmount
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.