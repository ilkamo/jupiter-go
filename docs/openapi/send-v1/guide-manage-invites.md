> ## Documentation Index
> Fetch the complete documentation index at: https://developers.jup.ag/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage Invites (Beta)

> Query pending and historical Send invites for a wallet address.

This guide covers how to retrieve pending and historical Send invites from the sender's perspective using the `/send/v1/pending-invites` and `/send/v1/invite-history` GET endpoints.

## Overview

**Step 1**

Get pending invites.

**Step 2**

Get invite history.

> [!NOTE]
>
> **NOTE**
>
> Both of the following endpoints only returns the invites that are set up by the sender and not from the perspective of the recipient.
>
> * Pending invites: Invites created by the sender that are not yet expired and can be clawback/claimed.
> * Invite history: Invites created by the sender and is either claimed, clawback, or expired. (You can also pass in a Recipient pubkey to get their history)

> [!TIP]
>
> **TIP**
>
> Depending on how you have set up to allow connection of wallets, either via [Jupiter Mobile Adapter](https://developers.jup.ag/docs/tool-kits/wallet-kit/jupiter-mobile-adapter.md) for QR code login, wallet extensions, or any other methods, you will need to handle the passing in of their pubkey to the API to get the necessary data.

## Get Pending Invites

```js theme={null}
const pendingInvites = await (
  await fetch(
    `https://api.jup.ag/send/v1/pending-invites?address=${pubkey}`,
    {
      headers: {
        'x-api-key': 'your-api-key',
      },
    }
  )
).json();
```

## Get Invite History

```js theme={null}
const inviteHistory = await (
  await fetch(
    `https://api.jup.ag/send/v1/invite-history?address=${pubkey}`,
    {
      headers: {
        'x-api-key': 'your-api-key',
      },
    }
  )
).json();
```


This documentation is built and hosted on [Mintlify](https://mintlify.com), a developer documentation platform.