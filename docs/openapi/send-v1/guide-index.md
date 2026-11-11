> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Jupiter Send API Overview

> Overview of Jupiter Send API for gifting, paying, or onboarding users even without a wallet

The Jupiter Send API lets you send tokens to any user — even if they don't have a connected wallet. Recipients claim their tokens through Jupiter Mobile. The API supports crafting send transactions, managing invites, and clawback of unclaimed sends.

Send is the perfect onboarding tool to gift, pay, or onboard anyone in seconds - even if they don't have a wallet.

* Send any token - SOL, USDC or memecoins.
* Send to a new user without a wallet, existing user or anyone.
* No fees to send or claim, only network transaction fees required.
* Use Jupiter Mobile seamlessly - even if you sent non-SOL tokens, Ultra provides Gasless Support that pays for swap transaction fees.

## About

Send API provides more opportunities for potential users to be onboarded from other websites, apps, or any where else!

* API only supports creating Send and Clawback transactions.
* Claiming needs to be done via Jupiter Mobile only.
* You can gamify the experience post-claim once they are onboarded!

## Jupiter Mobile Adapter

To maximize users experience, once your users have claimed via Jupiter Mobile, they can use the app to continue their journey on your app or other use cases. This can be done via [Jupiter Mobile Adapter](https://developers.jup.ag/docs/tool-kits/wallet-kit/jupiter-mobile-adapter.md), allowing Jupiter Mobile users to simply use the app to scan a QR code to login, they can utilize their wallets on Jupiter Mobile across any platform.

## FAQ

> [accordion] How to does claim work?
>
> * The invite code can be in the format of a link or a QR code.

> [accordion] How to claim via API or on my own app?
>
> * No, Send claims should be done in Jupiter Mobile.

> [accordion] Can I get my funds back?
>
> * Send is end-to-end self-custodial, where if recipient never claims, the invite code becomes invalid and your tokens are sent back to you upon expiry.
> * Or use the clawback endpoint via the API to create the clawback transaction.


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.