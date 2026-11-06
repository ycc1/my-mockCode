# KYCTest - Unico IDCloud UAT

KYCTest is a local Go web server with a separate browser frontend and backend API. It starts Unico's Mexico UAT flow, which includes liveness, identity verification, fraud-risk classification, and document capture. After Unico approves and finishes the process, the UI fetches the returned CURP, birth date, and full name for review and editing. It does not simulate approval.

## Run on Windows

Requirements: Go 1.22 or later, Node.js, and npm.

1. Open PowerShell in the `KYCTest` directory and run `.\start.ps1`. It installs the locked SDK dependencies, builds the browser SDK bundle, and starts the Go server. On first launch it creates `.env` from `.env.example`; existing PowerShell environment variables take precedence.
2. Save the **UAT** RSA PEM private key supplied by Unico as `KYCTest\secrets\unico-uat.pem`. Do not commit or share the private key.
3. Open <https://localhost:8080>. The server creates a self-signed development certificate in `KYCTest\certs` on first launch. Import `localhost.cert.pem` into the Current User **Trusted Root Certification Authorities** certificate store (for example, through `certmgr.msc`) and restart the browser if needed. Do not install the private key into the trust store.
4. Register `https://localhost:8080` as an allowed HTTPS domain with Unico before starting the verification journey. Configure `UNICO_CALLBACK_URI` to the matching HTTPS return URL if the callback is used by the configured flow.

The server starts without the PEM so the UI can be opened, but real Unico operations remain disabled until a valid UAT key is configured. The account and tenant defaults in `.env.example` come from the supplied UAT configuration. The documented IDCloud API authenticates with the service-account issuer and its RSA private key; it does not use a separate key-ID header. `go run .` can be used after running `npm ci` and `npm run build`.

Run the backend tests with:

```powershell
go test ./...
```

## Flow and records

- `POST /api/kyc/process` accepts an explicit consent confirmation and the user's CURP, and creates an `id_r2-idgov-r2` process through `POST /client/v1/process` with `person.duiType=DUI_TYPE_MX_CURP` and `person.duiValue` set to the CURP (the flow rejects requests without person data). The CURP is forwarded to Unico and is not stored in the pending record. No identity document is requested before the Unico-hosted journey.
- The Go backend signs a short-lived RS256 JWT assertion and exchanges it at the UAT OAuth token endpoint. No private key or access token is sent to the browser.
- The frontend opens the Unico-hosted verification page (`userRedirectUrl`, validated as an HTTPS `unico.app` URL) in a new browser tab; the tab is opened inside the click handler so pop-up blockers allow it. The original page shows a waiting screen and polls `GET /api/kyc/process/{recordId}` for up to two minutes; "Reopen verification tab" reopens the saved link. Open always shows the Start screen; if its saved record is still pending, Continue opens it again in a new tab.
- The waiting screen includes a demo-only button for exercising the waiting-screen transition. It does not simulate or save an approved result.
- Because there is no SDK callback, KYCTest relies on polling. Only after an approved, finished process does `GET /api/kyc/process/{recordId}/identity` return the extracted identity fields for review. `PATCH /api/kyc/process/{recordId}` saves the user's validated CURP, birth date, and full name.
- `GET /api/kyc/records` remains available for backend audit records; verification records are not shown in the UI. The record summaries omit the journey token and identity fields. `GET /healthz` is the server health check.
- KYCTest persists process status, the journey token required to resume, and identity fields returned by Unico or subsequently edited by the user in `data\records.json`. It does not store the original document, private key, or access token. The journey token and identity fields are sensitive; keep the local data directory protected.

The local record file is created with restrictive permissions where supported and is excluded from Git. The app binds to `localhost:8080` by default and is intended for local UAT testing, not as a production deployment or public-facing KYC service.

## Configuration

`start.ps1` loads `.env` into the process environment. The backend also accepts these environment variables directly:

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `localhost:8080` | Local HTTPS Go server address |
| `KYCTEST_DATA_DIR` | `./data` | Local record storage |
| `KYCTEST_TLS_CERT_FILE` | `./certs/localhost.cert.pem` | Local development certificate |
| `KYCTEST_TLS_KEY_FILE` | `./certs/localhost.key.pem` | Local development certificate key |
| `UNICO_ACCOUNT` | `megwin_prod_` | Unico service-account name |
| `UNICO_TENANT_ID` | Supplied UAT tenant ID | Issuer tenant |
| `UNICO_PRIVATE_KEY_FILE` | `./secrets/unico-uat.pem` | UAT RSA private key path |
| `UNICO_API_BASE` | `https://api.idcloud.uat.unico.app` | Unico IDCloud UAT API |
| `UNICO_TOKEN_URL` | `https://identityhomolog.acesso.io/oauth2/token` | UAT OAuth token endpoint |
| `UNICO_FLOW` | `id_r2-idgov-r2` | Mexico onboarding flow |
| `UNICO_CALLBACK_URI` | `https://localhost:8080/` | HTTPS return URL for the journey |

The configured UAT onboarding flow is `id_r2-idgov-r2`. Step 4 extracts and presents the identity fields available in the Unico process response; field availability depends on the configured recipe and captured document. The local callback is intended only for local browser testing.

## API documentation

- [Onboarding flow](https://developer.unico.io/zh-CN/products/sign-up/onboarding/index)
- [Non-document verification](https://developer.unico.io/zh-CN/products/sign-up/non-document-verification/index)
- [Available flows](https://developer.unico.io/zh-CN/developers/start/flows)
- [Authentication](https://developer.unico.io/zh-CN/developers/start/authentication)
- [Create process](https://developer.unico.io/zh-CN/developers/api-reference/post-processes)
- [Get process](https://developer.unico.io/zh-CN/developers/api-reference/get-process)
