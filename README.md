# zerotier-manager-bot

Simple telegram bot for managing zerotier-one-based VPN.

The app is now in development state and is not 100% documented.
Just look into the code to see how it works exactly.

###When to use:
You have a VPN based on ZeroTierOne and have to add/remove members frequently,
but it's inconvenient to visit zero tier central website every time?
Want more people to be able to add/remove members in your network, but don't have paid subscription?
You can use this bot to handle both cases.

###How to use:
- Create a telegram bot
- Allow incoming connections on some these ports: 443, 80, 88 and 8443 (telegram webhooks do not support others)
- Get a TLS certificate, e.g, self-signed. Here's a guide https://core.telegram.org/bots/webhooks#a-certificate-where-do-i-get-one-and-how.
- Find out your telegram user id (number)
- Create a config file such as the following example:
```yaml
token: 'your_telegram_bot_token_here'
web_hook_url: 'https://your.domain.name/your_webhook_uri'
web_hook_cert: 'YourCert.pem'
web_hook_key: 'YourPrivateKey.key'
listen_addr: '0.0.0.0' # Your server's IP address for webhook listening
port: 443 # port to listen webhooks on
zt_token: 'your_zerotier_zentral_api_token' # you can generate one in profile's settings
zt_network: "FFFFFFFFFFFFFFFF" # your ZeroTier network id (16 hexadecimal digits)
admin_id: 0 # telegram user id of admin
ops_file: "ops.txt" # file where to store list of server operators
```
- Run compiled executable (consider you named it zmanbot):
`./zmanbot --config=your_config.yml`
- Just do Ctrl+C to stop bot.

###How it works:
- Bot listens for updates on webhook (only message updates are being received; all pending messages are to be dropped)
- Bot ignores all non-command messages
- There are the only one admin determined in config file
    - Config file is the only way to set admin
    - Admin cannot be changed from application runtime
    - Admin can add and remove operators by telegram user id (`/op` and `/deop` respectively)
- Only operators and higher can use commands (except `/start`, that is available for all as it tells user id)
- Try `--help` flag to see command's help
