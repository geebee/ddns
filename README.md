## Dynamic DNS

A daemon and serverless worker to be your own Dynamic DNS service using Cloudflare.

### Daemon

The daemon will ensure that the configured record exists with the current external IP, and will keep it updated. It will not actually update the DNS record if the existing value is still valid.

#### Configuration

Configuration is done exclusively via the environment, there is no config file and no flags. However, if `./.env` is present when starting, its values will be read and set into the process' environment at startup, overriding everything else.

|Key                 |Required|Default|Description                                                                                        |
|:------------------:|:------:|:-----:|:--------------------------------------------------------------------------------------------------|
|`CLOUDFLARE_API_KEY`|  Yes   | None  |A Cloudflare API key with at least the `DNS:Edit` permission for the zone specified by `DNS_DOMAIN`|
|`DNS_DOMAIN`        |  Yes   | None  |The domain name to create the dynamic DNS record in                                                |
|`DNS_HOST`          |  Yes   | None  |The domain name to create the dynamic DNS record in                                                |
|`IP_LOOKUP_URL`     |  Yes   | None  |The URL of a service that will return the caller's external IP address in `text/plain`             |
|`REFRESH_INTERVAL`  |  No    | `24h` |The interval at which to check for an updated external IP address                                  |

### Worker

A serverless worker function written in Javascript that responds to the caller as follows:

|Method   |Path    |Response Type      |Status|Example Response        |
|:-------:|:------:|:-----------------:|:----:|:-----------------------|
|`GET`    |`/`     |`text/plain`       |`200` |`127.0.0.1`             |
|`GET`    |`/json` |`application/json` |`200` |`{"ip":"127.0.0.1"}`    |
|`GET`    |`*`     |`text/plain`       |`404` |`404 Not Found`         |
|non-`GET`|`*`     |`text/plain`       |`405` |`405 Method Not Allowed`|

#### Configuration

Configuration is done via the `package.json` file, and the `wrangler.toml` file. In order to avoid committing the cloudflare account ID, it is omitted from the configuration and should instead be specified with the `CF_ACCOUNT_ID` environment variable set when calling `wrangler` commands.

#### Deployment

Set up `wrangler` as per: https://developers.cloudflare.com/workers/get-started/guide
Run: `CF_ACCOUNT_ID="<cloudflare account ID>" wrangler publish`
