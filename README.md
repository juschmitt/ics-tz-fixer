# ICS Timezone Fixer

Small Go-based proxy that fetches an Outlook ICS feed, normalizes Outlook TZIDs using the embedded `VTIMEZONE` rules, and republishes a corrected ICS feed that Google Calendar can subscribe to.

## Usage

1. Deploy to Vercel.
2. Subscribe Google Calendar to:

```
https://<your-vercel-app>.vercel.app/api/calendar?url=<OUTLOOK_ICS_URL>
```

Optional override:

```
https://<your-vercel-app>.vercel.app/api/calendar?url=<OUTLOOK_ICS_URL>&tz_hint=Europe/Berlin
```

## Notes

- Only `outlook.office365.com` is allowed as the source host.
- If a `VTIMEZONE` signature matches a known IANA zone, the TZID is replaced.
- If a signature is unknown, the original TZID is left untouched.
- `VTIMEZONE` blocks are stripped after normalization.
- The handler returns `text/calendar` and streams the transformed content.

## Testing

```
go test ./...
```
