# Slot monitor

`cmd/slot-monitor` polls Playtomic availability for configured clubs and time
windows and sends a Telegram message when **new** matching padel slots appear.
It runs as a GitHub Actions workflow (`.github/workflows/slot-monitor.yml`) every
10 minutes.

Padel only: the sport is hard-coded, there is no sport option.

## What a run does

1. Logs in with `PLAYTOMIC_EMAIL` / `PLAYTOMIC_PASSWORD` (always a fresh login,
   no token caching).
2. For each club: fetches its courts (`GetResources`) and, per candidate day,
   availability (`GetAvailability`).
3. Keeps slots that pass the filters and fall fully inside a watch window.
4. Compares against the seen-slots state; sends Telegram only for slots not seen
   before; records them.
5. Exits green even when nothing is found — only real errors (bad config, auth,
   every club failing) make the run red.

## Configuration — repository variable `MONITOR_CONFIG`

The whole config is a JSON document stored in the **repository variable**
`MONITOR_CONFIG` (Settings → Secrets and variables → Actions → Variables). It is
written to `config.json` at run time, so changing it needs no commit. It lists
only clubs — court names and indoor/outdoor are resolved from the API at run
time.

```json
{
  "timezone": "Europe/Berlin",
  "look_ahead_days": 7,
  "clubs": [
    { "tenant_id": "a8b0e7b9-7db0-4c45-8569-28bf775e9208", "name": "halle11 (Berg)", "url": "https://playtomic.com/clubs/tennishalle-berg" },
    { "tenant_id": "a51996c7-343f-4ff5-b6c9-902fbf06ba7a", "name": "Padelmotion (Weingarten)", "url": "https://playtomic.com/clubs/padelmotion" }
  ],
  "watch_windows": [
    { "days": ["mon", "tue", "wed", "thu"], "start": "17:30", "end": "21:00" },
    { "days": ["sat", "sun"], "start": "09:00", "end": "22:00" }
  ],
  "filters": {
    "durations": [90],
    "court_type": "",
    "court_size": "",
    "include_court_names": [],
    "exclude_court_names": []
  }
}
```

| Field | Meaning |
| --- | --- |
| `timezone` | Your target timezone (IANA name). All windows and message times use it. Default `Europe/Berlin`. |
| `look_ahead_days` | How many days ahead to check. Default 7. |
| `clubs[].tenant_id` | Required. The Playtomic tenant id of the club. |
| `clubs[].name`, `clubs[].url` | Optional, only for nicer Telegram messages. |
| `watch_windows[]` | Per-weekday time spans. A slot must fit **fully** inside a window. Use several entries for weekday/weekend differences. |
| `filters.durations` | Allowed slot lengths in minutes. Empty = any. |
| `filters.court_type` | `""`, `"indoor"` or `"outdoor"`. |
| `filters.court_size` | `""`, `"single"` or `"double"`. |
| `filters.include_court_names` / `exclude_court_names` | Case-insensitive substring match on the court name. Empty include = all. |

Days are `mon tue wed thu fri sat sun`.

## Secrets

Add these under Settings → Secrets and variables → Actions → Secrets:

- `PLAYTOMIC_EMAIL`, `PLAYTOMIC_PASSWORD` — your Playtomic login.
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` — the bot and target chat. Without
  them the run still works and logs what it would have sent.

## State

Already-notified slots are remembered in `state/seen.json`, persisted through the
GitHub Actions cache (no commits, no token needed). If the cache is evicted a
slot may be reported a second time.

## Timezone note

The Playtomic app API reports slot times in **UTC**; the monitor converts them to
your `timezone` for filtering and display. Verify on the first real hit that the
message time matches what playtomic.com shows.

## Local run

```bash
export PLAYTOMIC_EMAIL=... PLAYTOMIC_PASSWORD=...
export TELEGRAM_BOT_TOKEN=... TELEGRAM_CHAT_ID=...   # optional
go run ./cmd/slot-monitor -config config.json -state state/seen.json
```
