# UFC


[![CI](https://github.com/wajeht/ufc/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/wajeht/ufc/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/wajeht/ufc/blob/main/LICENSE) 
[![Open Source Love svg1](https://badges.frapsoft.com/os/v1/open-source.svg?v=103)](https://github.com/wajeht/ufc)

Subscribe to UFC events in your calendar app.

Scrapes upcoming events from ufc.com, generates an `.ics` calendar file, and serves it via a simple web server.

## Endpoints

| Endpoint       | Description           |
| -------------- | --------------------- |
| `/`            | upcoming events       |
| `/events.ics`  | calendar subscription |
| `/events.json` | raw event data        |
| `/health`      | health check          |

## License

Distributed under the MIT License © [wajeht](https://github.com/wajeht). See [LICENSE](./LICENSE) for more information.
