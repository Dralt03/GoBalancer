# PowerShell script to run benchmarks for GoBalancer

Write-Host "--- Running Internal Go Benchmarks ---" -ForegroundColor Cyan
go test -bench . -benchmem ./internal/backend ./internal/balancer

Write-Host "`n--- Starting Mock Backend ---" -ForegroundColor Cyan
$BackendProcess = Start-Process go -ArgumentList "run cmd/mock_backend/main.go 8080" -PassThru -NoNewWindow
Start-Sleep -Seconds 2

Write-Host "`n--- Running Proxy with Mock Backend ---" -ForegroundColor Cyan
Write-Host "Note: This assumes the proxy is running and listening on :8080" -ForegroundColor Yellow

# Run the load tester against the proxy
go run cmd/load_tester/main.go -url http://localhost:8080 -c 20 -d 5s

$BackendProcess | Stop-Process
Write-Host "`nBenchmarks completed." -ForegroundColor Green
