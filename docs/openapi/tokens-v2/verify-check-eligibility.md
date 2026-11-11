> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Check Eligibility

> Check if a token is eligible for express verification and metadata updates



## OpenAPI

````yaml openapi-spec/tokens/v2/verification.yaml get /express/check-eligibility
openapi: 3.0.3
info:
  title: Jupiter Token Verification API
  version: 1.0.0
  description: API for express token verification and metadata updates via Jupiter VRFD
servers:
  - url: https://api.jup.ag/tokens/v2/verify
    description: Jupiter Token Verification API Endpoint
security: []
paths:
  /express/check-eligibility:
    get:
      tags:
        - Express Verification
      summary: Check express verification eligibility
      description: >
        Checks whether a token is eligible for express verification and metadata
        updates.

        If both `canVerify` and `canMetadata` are `false`, any submission will
        be rejected

        before payment is charged.
      parameters:
        - in: query
          name: tokenId
          schema:
            type: string
          required: true
          description: Token mint address to check eligibility for
          example: EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
      responses:
        '200':
          description: Eligibility check result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CheckEligibilityResponse'
              example:
                tokenExists: true
                isVerified: false
                canVerify: true
                canMetadata: false
                metadataError: >-
                  A premium metadata update request for this token already
                  exists
        '400':
          description: Invalid request parameters
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - ApiKeyAuth: []
components:
  schemas:
    CheckEligibilityResponse:
      type: object
      required:
        - tokenExists
        - isVerified
        - canVerify
        - canMetadata
      properties:
        tokenExists:
          type: boolean
          description: Whether the token exists on-chain
        isVerified:
          type: boolean
          description: Whether the token is already verified
        canVerify:
          type: boolean
          description: Whether a verification request can be submitted
        canMetadata:
          type: boolean
          description: Whether a metadata update can be submitted
        verificationError:
          type: string
          description: >-
            Reason verification is not allowed (only present when there is a
            specific error)
        metadataError:
          type: string
          description: >-
            Reason metadata update is not allowed (only present when there is a
            specific error)
    ErrorResponse:
      type: object
      properties:
        status:
          type: number
        message:
          type: string
        error:
          type: string
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: x-api-key
      description: Get API key via https://developers.jup.ag/portal

````

This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.