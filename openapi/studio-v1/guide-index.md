> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Jupiter Studio Overview

> Jupiter Studio provides APIs for token creation with flexible bonding curves, LP fee management, vesting schedules, and dedicated token pages on jup.ag.

Studio is built for culture architects who want:

* Aggressive experimentation.
* Tools for growth and alignment.
* Collaborative and supportive vibe culture between Studio projects.

## About

Studio is a powerful playground equipped with a suite of tools for creators. Each feature is strategic towards how creators might want to customize to fit their needs - like flexible bonding curves, custom vesting schedules, and selectable quote mints to encode your vision.

**Features**

* LP Fees: 50% before AND after graduation.
* LP Locking: Optional 50% of the graduated LP unlocks after 1 year.
* Vested Tokens: 0 - 80% of token supply, with optional vesting schedule and cliff.
* Flexible parameters: Quote mint, Market cap bonding, etc.
* Other helpful tools: Anti-sniper suite, Lp Locking.

**Dedicated Studio Token Page**

Apart from the strategic levers, start rallying your community with the dedicated Studio page with seamless content integration with jup.ag's token page.

* Dedicated Studio page for each token.
* Content from Studio shows up in jup.ag's token page.

> [!NOTE]
>
> **READINGS**
>
> * Design intentions: [https://x.com/9yointern/status/1940431614103937517](https://x.com/9yointern/status/1940431614103937517)
> * Launch post: [https://x.com/jup\_studio/status/1940620377602011566](https://x.com/jup_studio/status/1940620377602011566)
> * General FAQ: [https://docs.jup.ag/user-docs/launch/studio](https://docs.jup.ag/user-docs/launch/studio)

## FAQ

> [accordion] Why is my Studio token page showing 'Token not found'?
>
> * In order for us to track and store your token information, header image or token description, you **must** send your signed transaction from the `create_tx` endpoint to the `submit` endpoint.
> * This will allow us to store your token into our database and reflect it as a Studio token on our frontend.
> * If you submit the transaction on your own or some other way, the token will not have a dedicated Studio page.

> [accordion] What do I do with the presigned URLs?
>
> * Those URLs are for you to upload your token's metadata and image to a static endpoint, which will be in the token's URI metadata onchain.
> * You are required to make a PUT request to those endpoints, [you can refer to this section on the usage](https://developers.jup.ag/docs/studio/create-token.md#token-metadata).
> * If you do not upload your token image and metadata to this endpoint, your token will not have any image/metadata reflected onchain.

> [accordion] What is the rate limit of Studio API?
>
> * Free usage rate limit: 100 requests per 5 minutes
> * For paid rate limit: 10 requests per 10 seconds (for all Tiers)


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.