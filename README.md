# qr_generation

PoC Go service and library for generating UPI QR PNG cards and batch ZIP downloads.

## Scope

This first version does three things:

1. Accepts a `upi://pay?...` URI and generates a QR code server-side.
2. Composes the QR into a fixed branded PNG card with merchant text, description, and optional logo.
3. Accepts multiple items and returns a ZIP archive of PNG files with sanitized unique names.

The renderer is native Go image composition. It does not use HTML templates or browser screenshot tooling.

## Project layout

- `qrgen`: reusable public package for rendering one PNG or one ZIP archive.
- `cmd/server`: thin HTTP server for demo and integration.
- `internal/upi`: UPI URI validation.
- `internal/render`: fixed card layout and PNG rendering.
- `internal/batch`: ZIP orchestration and filename generation.

## Run locally

```bash
go mod tidy
go run ./cmd/server
```

The server listens on `:8080` by default. Override with `QRGEN_ADDR`.

## HTTP API

### `POST /v1/render`

Returns `image/png`.

Example request:

```json
{
  "upi_uri": "upi://pay?pa=simbapvt7140@idfcbank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201234",
  "merchant_name": "Simba Pvt Ltd",
  "merchant_upi_id": "simbapvt7140@idfcbank",
  "description": "Invoice 1234",
  "provider_name": "IDFC First Bank",
  "payer_name": "Mohammad Khizer Khan",
  "qr_generator": "piglig",
  "accent_color": "#9f1a1a"
}
```

Example:

```bash
curl -X POST http://localhost:8080/v1/render \
	-H 'Content-Type: application/json' \
	-d '{
		"upi_uri":"upi://pay?pa=simbapvt7140@idfcbank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201234",
		"merchant_name":"Simba Pvt Ltd",
		"merchant_upi_id":"simbapvt7140@idfcbank",
		"description":"Invoice 1234",
		"provider_name":"IDFC First Bank"
	}' \
	--output upi-card.png
```

### `POST /v1/batch`

Returns `application/zip`.

```json
{
  "items": [
    {
      "upi_uri": "upi://pay?pa=simbapvt7140@idfcbank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201234",
      "merchant_name": "Simba Pvt Ltd",
      "payer_name": "Alice"
    },
    {
      "upi_uri": "upi://pay?pa=simbapvt7140@idfcbank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201235",
      "merchant_name": "Simba Pvt Ltd",
      "payer_name": "Bob"
    }
  ]
}
```

Example:

```bash
curl -X POST http://localhost:8080/v1/batch \
	-H 'Content-Type: application/json' \
	-d @batch.json \
	--output upi-cards.zip
```

## Library usage

```go
svc := qrgen.NewService()
pngData, err := svc.RenderPNG(qrgen.CardRequest{
		UPIURI:       "upi://pay?pa=merchant@bank&pn=Shop",
		MerchantName: "Shop",
})
```

## Request fields

- `upi_uri`: required, must start with `upi://pay` and include `pa`.
- `merchant_name`: required unless present as `pn` inside the UPI URI.
- `merchant_upi_id`: optional override; defaults to the `pa` parameter.
- `description`: optional; defaults to `tn` if present, otherwise a generic prompt.
- `provider_name`: optional footer text when no logo is provided.
- `payer_name`: optional; used on the card and in batch filenames.
- `logo_base64`: optional PNG or JPEG logo as raw base64 or data URI.
- `qr_generator`: optional QR backend selector. Supported values: `skip2` (default), `yeqown`, and `piglig`.
- `background_color`, `accent_color`, `text_color`: optional hex colors in `#RRGGBB` format.

## Validation

- Non-UPI schemes are rejected.
- URIs must target `upi://pay`.
- The `pa` parameter is required.
- Invalid base64 logos or invalid colors return `400` responses from the HTTP layer.
- `yeqown` and `piglig` use a built-in static QR center logo from the codebase.

## Tests

```bash
go test ./...
```

```go
go run ./scripts/generate_batch_payload.go -n 5000 -out testdata/batch_payload_5000.json
```
