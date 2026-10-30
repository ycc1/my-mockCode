# KYCTest - Unico IDCloud UAT

KYCTest is a local Go web server with a separate browser frontend and backend API. It starts Unico's Mexico UAT flow, which includes liveness, identity verification, fraud-risk classification, and document capture. After Unico finishes, the UI displays CURP, birth date, and full name returned by Unico for review and editing. It does not simulate approval.

## Run on Windows

Requirements: Go 1.22 or later.

1. Open PowerShell in the `KYCTest` directory. Run `.\start.ps1` (or `go run .` to start directly). On first launch, `start.ps1` creates `.env` from `.env.example`; Go loads `.env` automatically, while environment variables already set in PowerShell take precedence.
2. Save the **UAT** RSA PEM private key supplied by Unico as `KYCTest\secrets\unico-uat.pem`. Do not commit or share the private key.
3. Check `.env` and set `UNICO_CALLBACK_URI` to the URL where the browser should return after Unico's hosted journey. The default is suitable only when running locally; a deployed environment should use its reachable HTTPS address.
4. Open <http://localhost:8080>.

The server starts without the PEM so the UI can be opened, but real Unico operations remain disabled until a valid UAT key is configured. The account and tenant defaults in `.env.example` come from the supplied UAT configuration. The documented IDCloud API authenticates with the service-account issuer and its RSA private key; it does not use a separate key-ID header.

Run the backend tests with:

```powershell
go test ./...
```

## Flow and records

- `POST /api/kyc/process` accepts an explicit consent confirmation and creates a `multiaccounts_idgov` process through `POST /client/v1/process`. No CURP or identity document is requested from the user before the Unico-hosted journey.
- The Go backend signs a short-lived RS256 JWT assertion and exchanges it at the UAT OAuth token endpoint. No private key or access token is sent to the browser.
- After the hosted journey, `GET /api/kyc/process/{recordId}` retrieves the result and extracted identity fields from Unico. `PATCH /api/kyc/process/{recordId}` validates and saves the user's edited CURP, birth date, and full name.
- `GET /api/kyc/records` lists local audit records. `GET /healthz` is the server health check.
- KYCTest persists process status and the identity fields returned by Unico or subsequently edited by the user in `data\records.json`. It does not store the original document, private key, or access token. The identity fields are sensitive; keep the local data directory protected.

The local record file is created with restrictive permissions where supported and is excluded from Git. The app binds to `127.0.0.1` by default and is intended for local UAT testing, not as a production deployment or public-facing KYC service.

## Configuration

`start.ps1` loads `.env` into the process environment. The backend also accepts these environment variables directly:

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `127.0.0.1:8080` | Local Go server address |
| `KYCTEST_DATA_DIR` | `./data` | Local record storage |
| `UNICO_ACCOUNT` | `megwin_prod_` | Unico service-account name |
| `UNICO_TENANT_ID` | Supplied UAT tenant ID | Issuer tenant |
| `UNICO_PRIVATE_KEY_FILE` | `./secrets/unico-uat.pem` | UAT RSA private key path |
| `UNICO_API_BASE` | `https://api.idcloud.uat.unico.app` | Unico IDCloud UAT API |
| `UNICO_TOKEN_URL` | `https://identityhomolog.acesso.io/oauth2/token` | UAT OAuth token endpoint |
| `UNICO_FLOW` | `multiaccounts_idgov` | Mexico onboarding flow with liveness, identity, fraud-risk classification, document capture, face matching, RENAPO, and multi-account detection |
| `UNICO_CALLBACK_URI` | `http://localhost:8080/` | Return URL for the hosted journey |

Unico support confirmed that `multiaccounts_idgov` is the UAT onboarding flow enabled for the `megwin_uat` account. It includes liveness, identity verification, fraud-risk classification, document capture and face matching, RENAPO verification, and multi-account detection. Step 4 extracts and presents the identity fields available in the Unico process response; field availability depends on the configured recipe and captured document. The local callback is intended only for local browser testing.

## API documentation

- [Onboarding flow](https://developer.unico.io/zh-CN/products/sign-up/onboarding/index)
- [Non-document verification](https://developer.unico.io/zh-CN/products/sign-up/non-document-verification/index)
- [Available flows](https://developer.unico.io/zh-CN/developers/start/flows)
- [Authentication](https://developer.unico.io/zh-CN/developers/start/authentication)
- [Create process](https://developer.unico.io/zh-CN/developers/api-reference/post-processes)
- [Get process](https://developer.unico.io/zh-CN/developers/api-reference/get-process)
