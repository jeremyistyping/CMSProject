# Bugfix: PR Verification 404 Error

## Problem
Error 404 saat memanggil endpoint `POST /api/v1/purchase-requests/:id/verify` dari frontend.

```
AxiosError: Request failed with status code 404
at async Object.verifyPR (src/services/cbsService.ts:58:9)
at async handleSubmit (src/components/cost-control/PRVerificationModal.tsx:190:13)
```

## Root Cause
Endpoint handler `VerifyPurchaseRequest` sudah diimplementasikan di `CBSController`, tetapi **tidak terdaftar di routing** (`backend/routes/routes.go`).

## Solution
Menambahkan 2 route yang hilang ke Purchase Request group:

```go
// CBS Verification routes
pr.POST("/:id/verify", permMiddleware.CanApprove("purchases"), cbsController.VerifyPurchaseRequest)
pr.GET("/:id/cbs-mappings", permMiddleware.CanView("purchases"), cbsController.GetPRCBSMappings)
```

## Files Changed
- `backend/routes/routes.go` - Menambahkan routing untuk PR verification dan CBS mappings

## Testing
1. Restart backend server
2. Test PR verification dari frontend:
   - Buka PR yang pending verification
   - Klik "Verify & Map to CBS"
   - Pilih CBS nodes dan submit
   - Seharusnya berhasil tanpa error 404

## Related Endpoints
- `POST /api/v1/purchase-requests/:id/verify` - Verify PR dan simpan CBS mappings
- `GET /api/v1/purchase-requests/:id/cbs-mappings` - Get CBS mappings untuk PR

## Permissions Required
- Verify: `CanApprove("purchases")`
- Get Mappings: `CanView("purchases")`
