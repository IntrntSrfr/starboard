# Starboard Privacy Policy

Effective date: October 1, 2026

This policy covers the Starboard Discord bot operated by IntrntSrfr,
application ID `622103293770072065`. Independently hosted copies of the
software are covered by their operators' policies.

For privacy questions or data-removal requests, email the operator at
[intrntsrfr@gmail.com](mailto:intrntsrfr@gmail.com).

## What Starboard does with your data

Starboard highlights messages when they receive the number of star reactions
configured by a server administrator. It processes message text, images and
attachment links, author names and avatars, reactions, and server, channel,
message, and user IDs. It also processes message edits and deletions, server
settings, and commands or messages sent to the bot.

A highlight includes selected message content, attribution, the star count,
and a link to the original. It is posted in the same server's configured
starboard channel. Anyone who can view that channel can see the highlight,
even if they cannot view the original channel. Other members can nominate a
message by starring it; the author does not need to invoke the bot.

The bot uses data to provide these features and diagnose errors. Its features
do not include advertising, user profiling, or training AI models.

## What is stored and for how long

- **PostgreSQL database:** Server IDs, destination-channel settings, star
  thresholds, and mappings between original and highlighted message/channel
  IDs. The database does not store message bodies or attachment files. Records
  persist until removed by the bot's event handlers or the operator. There is
  no scheduled expiry or automatic database cleanup when the bot leaves a server.
- **Temporary memory:** The current framework caches up to 100 messages per
  channel, including messages that have not been starred. It also retrieves
  and caches server member information, including IDs, names, nicknames, and
  roles. Memory caches are cleared when the process exits; entries may also be
  removed or replaced while it runs. They do not have a fixed time-based expiry.
- **Operational logs:** PM2 stores the bot's process output on the hosting
  server. Logs include startup activity, command metadata such as user and
  channel IDs, and errors. Some error paths include full message, interaction,
  or server-event payloads, so logs can contain message content and member data.
  No PM2 or matching system logrotate policy is currently configured for these
  logs; they remain on the server until cleared.
- **Highlights on Discord:** Highlights remain until removed. The bot attempts
  to remove them when the source message is deleted, the star count drops below
  the threshold, or all reactions are removed. Missed events, downtime, and
  errors can prevent automatic removal. Contact the operator if a highlight
  needs to be removed manually.

The bot uses existing attachment and media links and does not maintain its own
archive of attachment files.

## Where data is handled

The bot and its PostgreSQL database are hosted on the operator's Microsoft
Azure server in Ireland. PM2 logs are stored on that server. Its disk uses
Azure-managed encryption at rest. Microsoft Azure provides the hosting
infrastructure; the operator uses the stored data to run and maintain the bot.

Starboard communicates with Discord to receive data and post highlights.
Discord handles information on its platform under its own
[Privacy Policy](https://discord.com/privacy).

## Your choices and deletion requests

You can email [intrntsrfr@gmail.com](mailto:intrntsrfr@gmail.com) to request
access to, correction of, or deletion of data relating to you. For
example, you can ask to remove a starboard copy of your message, its database
mapping, or related information in logs or cached data under the operator's
control. Requests are handled manually. The operator may ask for your Discord
user ID and relevant message links to locate the data and verify the request.

Starboard currently has no individual opt-out command or user exclusion list.
Removing a highlight does not prevent future messages from being highlighted.
Server administrators can restrict the bot's channel permissions or remove it
to limit future access. Removing the bot does not itself erase existing
highlights, database records, or logs.

## Policy updates

Changes to this policy will be published here with an updated effective date.
