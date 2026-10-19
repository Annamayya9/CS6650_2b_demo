param([string]$BaseUrl = "http://localhost:8080/v1", [string]$ApiKey = "local-demo-key")
$ErrorActionPreference = "Stop"
# curl.exe avoids the Windows PowerShell curl alias. Requires ENABLE_TEST_ERRORS=true.
$cases = @(
 @{Name="GET 200"; Method="GET"; Path="/products/1"; Status=200},
 @{Name="GET 404"; Method="GET"; Path="/products/99"; Status=404},
 @{Name="POST 204"; Method="POST"; Path="/products/1/details"; Status=204; File="product.json"},
 @{Name="POST 400"; Method="POST"; Path="/products/1/details"; Status=400; File="invalid-product.json"},
 @{Name="POST 404"; Method="POST"; Path="/products/99/details"; Status=404; File="missing-product.json"},
 @{Name="GET 500"; Method="GET"; Path="/products/1"; Status=500; Fault=$true},
 @{Name="POST 500"; Method="POST"; Path="/products/1/details"; Status=500; File="product.json"; Fault=$true}
)
foreach ($case in $cases) {
 $curlArgs = @("-sS", "-i", "-w", "`nSTATUS:%{http_code}", "-X", $case.Method, "-H", "X-API-Key: $ApiKey")
 if ($case.File) { $curlArgs += @("-H", "Content-Type: application/json", "--data-binary", "@$PSScriptRoot/$($case.File)") }
 if ($case.Fault) { $curlArgs += @("-H", "X-Test-Error: 500") }
 $curlArgs += "$BaseUrl$($case.Path)"
 $response = & curl.exe @curlArgs
 if ($LASTEXITCODE -ne 0) { throw "curl failed: $($case.Name)" }
 $text = $response -join "`n"
 Write-Output "$($case.Name):`n$text`n"
 if ($text -notmatch "STATUS:$($case.Status)$") { throw "Unexpected response: $($case.Name)" }
}
Write-Output "PASS: all seven specified endpoint/status combinations."
