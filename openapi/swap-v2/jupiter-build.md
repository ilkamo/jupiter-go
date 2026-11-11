> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Build Transaction

> Get a swap quote with raw instructions for building custom transactions



## OpenAPI

````yaml openapi-spec/swap/v2/swap.yaml get /build
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
  /build:
    get:
      tags:
        - Swap
      summary: Build Transaction
      description: >
        Get a swap quote with raw instructions for building custom transactions.

        This is the advanced path for integrators who need full transaction
        control

        (add custom instructions, CPI, custom payer).


        You are responsible for assembling the transaction, signing, and sending
        via your own RPC.

        The `/execute` endpoint is not available for `/build` transactions.


        Routing is Metis only. No swap fees are charged by Jupiter.
      operationId: GetBuild
      parameters:
        - $ref: '#/components/parameters/InputMintParameter'
        - $ref: '#/components/parameters/OutputMintParameter'
        - $ref: '#/components/parameters/AmountParameter'
        - name: taker
          in: query
          required: true
          schema:
            type: string
          description: Public key of the user initiating the swap.
          example: GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ
        - $ref: '#/components/parameters/BuildSlippageBpsParameter'
        - name: mode
          in: query
          schema:
            type: string
            enum:
              - fast
          description: |
            Quoting mode. "fast" trades quote optimality for reduced latency.
        - name: dexes
          in: query
          schema:
            type: string
          description: >
            - Comma-separated list of DEXes to restrict routing to.

            - Labels are case sensitive (`SolFi`, not `solfi`). An unrecognised
            label is not a validation error: it returns `400 {"error": "No
            routes found"}`, indistinguishable from a pair with no liquidity.

            - Mutually exclusive with excludeDexes. Passing both returns `400
            {"error": "dexes and excludeDexes are mutually exclusive"}`.

            - Full list of DEXes here using the `/program-id-to-label` endpoint.
        - name: excludeDexes
          in: query
          schema:
            type: string
          description: |
            - Comma-separated list of DEXes to exclude.
            - Labels are case sensitive (`SolFi`, not `solfi`).
            - Mutually exclusive with dexes.
            - Full list of DEXes here using the `/program-id-to-label` endpoint.
        - name: platformFeeBps
          in: query
          schema:
            type: integer
            minimum: 0
            maximum: 10000
          description: |
            Platform fee in basis points. If positive, feeAccount is required.
          example: 50
        - name: feeAccount
          in: query
          schema:
            type: string
          description: >-
            Token account to collect platform fees. Required if platformFeeBps
            is positive.
        - name: maxAccounts
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 64
          description: Maximum number of accounts for the swap route. Defaults to 64.
        - name: maxBytes
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 1232
          description: >
            Target maximum transaction size in bytes, leaving room for your own
            instructions.


            - Sizing is deterministic, not an estimate: Metis builds and
            serializes the route to measure it. The swap it returns always fits
            your budget.

            - Counts the bare swap only. Anything you add on top is not counted,
            so your final transaction can still exceed it: request parameters
            (`payer`, `feeAccount`, `tipAmount`, destination accounts, close
            authority, cashback), your own instructions, extra signers, and your
            own lookup tables.

            - Set the budget below 1232 to leave room, then serialize your final
            transaction and lower `maxBytes` further if it still exceeds 1232.
            As a rough guide, each added account costs 1 byte if in your lookup
            table else 32, and each extra signer about 96 bytes.

            - v0 only (400 with `transactionVersion=1`).
        - name: payer
          in: query
          schema:
            type: string
          description: >-
            Account that pays transaction fees and rent. Defaults to taker when
            not passed in.
        - name: wrapAndUnwrapSol
          in: query
          schema:
            type: boolean
          description: Whether to wrap/unwrap SOL. Defaults to true.
        - name: destinationTokenAccount
          in: query
          schema:
            type: string
          description: |
            - SPL token account to receive output tokens.
            - Mutually exclusive with nativeDestinationAccount.
        - name: nativeDestinationAccount
          in: query
          schema:
            type: string
          description: |
            - Native SOL account to receive output.
            - Mutually exclusive with destinationTokenAccount.
        - name: blockhashSlotsToExpiry
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 300
          description: Number of slots until the blockhash expires. Defaults to 150.
        - name: tipAmount
          in: query
          schema:
            type: string
          description: >-
            SOL tip amount in lamports. Adds a tip instruction for use with
            /submit. The tip payer is the taker (or payer if specified).
          required: false
        - $ref: '#/components/parameters/BuildComputeUnitPricePercentileParameter'
        - name: forJitoBundle
          in: query
          schema:
            type: boolean
            default: false
          description: >
            Set to `true` if the transaction will be submitted as part of a Jito
            bundle.

            This excludes DEXes that are incompatible with Jito bundles.
        - name: transactionVersion
          in: query
          schema:
            type: string
            enum:
              - '0'
              - '1'
            default: '0'
          description: >
            The Solana transaction version to build for, `"0"` or `"1"`.
            Defaults to `"0"`.

            `"1"` returns v1 instructions: `computeBudgetInstructions` is empty
            and `addressesByLookupTableAddress` is `{}` (v1 inlines accounts, no
            lookup tables). Set the compute budget, including the returned
            `computeUnitPrice`, on the message config. Build with a v1-capable
            SDK (`@solana/kit` v8 or later).
          example: '1'
      responses:
        '200':
          description: Quote with raw swap instructions
          content:
            application/json:
              schema:
                type: object
                properties:
                  inputMint:
                    type: string
                  outputMint:
                    type: string
                  inAmount:
                    type: string
                  outAmount:
                    type: string
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
                      Price impact of the swap as a decimal ratio (e.g. "0.001"
                      = 0.1%)
                  routePlan:
                    type: array
                    items:
                      $ref: '#/components/schemas/RoutePlanStep'
                  computeBudgetInstructions:
                    type: array
                    items:
                      $ref: '#/components/schemas/Instruction'
                    description: >-
                      Compute unit price instruction (does not include compute
                      unit limit). Empty when transactionVersion is "1"; use
                      computeUnitPrice on the message config instead.
                  setupInstructions:
                    type: array
                    items:
                      $ref: '#/components/schemas/Instruction'
                    description: Pre-swap setup instructions (e.g. create ATAs)
                  swapInstruction:
                    $ref: '#/components/schemas/Instruction'
                  cleanupInstruction:
                    $ref: '#/components/schemas/Instruction'
                    nullable: true
                    description: Post-swap cleanup instruction
                  otherInstructions:
                    type: array
                    items:
                      $ref: '#/components/schemas/Instruction'
                  tipInstruction:
                    nullable: true
                    allOf:
                      - $ref: '#/components/schemas/Instruction'
                  addressesByLookupTableAddress:
                    type: object
                    additionalProperties:
                      type: array
                      items:
                        type: string
                    nullable: true
                    description: >-
                      Address lookup table mappings for v0 transactions. Empty
                      object ({}) when transactionVersion is "1" (v1 inlines
                      accounts, no lookup tables).
                  transactionVersion:
                    type: number
                    enum:
                      - 0
                      - 1
                    description: >-
                      The Solana transaction version these instructions are for,
                      0 or 1, matching the transactionVersion request parameter.
                  computeUnitPrice:
                    type: string
                    description: >-
                      Present when transactionVersion is "1". Compute unit price
                      in micro-lamports, to set on the message config.
                  blockhashWithMetadata:
                    type: object
                    properties:
                      blockhash:
                        type: array
                        items:
                          type: number
                        description: Blockhash as byte array
                      lastValidBlockHeight:
                        type: number
                      fetchedAt:
                        type: object
                        description: >-
                          When the blockhash was fetched, as a Unix timestamp
                          split into seconds and nanoseconds
                        properties:
                          secs_since_epoch:
                            type: number
                            description: Whole seconds since the Unix epoch
                          nanos_since_epoch:
                            type: number
                            description: Nanoseconds component within the second
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                type: object
                properties:
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
    BuildSlippageBpsParameter:
      name: slippageBps
      in: query
      schema:
        anyOf:
          - type: string
            enum:
              - rtse
          - type: number
            minimum: 0
            maximum: 10000
        default: 50
      description: Slippage tolerance in basis points. Defaults to 50.
    BuildComputeUnitPricePercentileParameter:
      name: computeUnitPricePercentile
      in: query
      schema:
        anyOf:
          - type: string
            enum:
              - medium
              - high
              - veryHigh
          - type: integer
            minimum: 0
            maximum: 10000
      description: >-
        Percentile of compute unit price to use. Named levels: "medium" (25th),
        "high" (50th), "veryHigh" (75th), or a number 0-10000 in basis points.
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
    Instruction:
      type: object
      properties:
        programId:
          type: string
          description: Program executing the instruction
        accounts:
          type: array
          items:
            type: object
            properties:
              pubkey:
                type: string
              isWritable:
                type: boolean
              isSigner:
                type: boolean
            required:
              - pubkey
              - isWritable
              - isSigner
        data:
          type: string
          description: Base64-encoded instruction data
      required:
        - programId
        - accounts
        - data
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