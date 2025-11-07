# SSOT Performance Monitoring Script
# Runs the performance monitoring system with various options

param(
    [switch]$Demo,
    [switch]$Dashboard,
    [switch]$Help,
    [int]$Duration = 300,  # Default 5 minutes
    [int]$Interval = 30    # Default 30 seconds
)

function Show-Help {
    Write-Host "📊 SSOT Performance Monitor" -ForegroundColor Cyan
    Write-Host "=============================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\run_performance_monitor.ps1 [options]" -ForegroundColor White
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -Demo              Run in demo mode (limited duration)" -ForegroundColor White
    Write-Host "  -Dashboard         Open web dashboard after starting monitor" -ForegroundColor White  
    Write-Host "  -Duration [seconds] Duration for demo mode (default: 300)" -ForegroundColor White
    Write-Host "  -Interval [seconds] Monitoring interval (default: 30)" -ForegroundColor White
    Write-Host "  -Help              Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\run_performance_monitor.ps1 -Demo" -ForegroundColor Green
    Write-Host "  .\run_performance_monitor.ps1 -Dashboard" -ForegroundColor Green
    Write-Host "  .\run_performance_monitor.ps1 -Demo -Duration 120" -ForegroundColor Green
    Write-Host ""
}

function Test-Prerequisites {
    Write-Host "🔍 Checking prerequisites..." -ForegroundColor Yellow
    
    # Check if Go is installed
    try {
        $null = Get-Command go -ErrorAction Stop
        $goVersion = go version
        Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
        return $false
    }
    
    # Check if we're in the correct directory
    if (-not (Test-Path "cmd/scripts/ssot_performance_monitor.go")) {
        Write-Host "❌ Performance monitor script not found. Please run from backend directory." -ForegroundColor Red
        return $false
    }
    
    # Check if config exists
    if (-not (Test-Path "config")) {
        Write-Host "❌ Config directory not found" -ForegroundColor Red
        return $false
    }
    
    Write-Host "✅ Prerequisites check passed" -ForegroundColor Green
    return $true
}

function Start-PerformanceMonitor {
    param(
        [bool]$IsDemoMode,
        [int]$MonitorDuration,
        [int]$MonitorInterval
    )
    
    Write-Host ""
    Write-Host "📊 Starting SSOT Performance Monitor" -ForegroundColor Cyan
    Write-Host "====================================" -ForegroundColor Cyan
    
    if ($IsDemoMode) {
        Write-Host "🎯 Demo Mode: Will run for $MonitorDuration seconds" -ForegroundColor Yellow
        Write-Host "⏱️  Monitoring interval: $MonitorInterval seconds" -ForegroundColor Yellow
    } else {
        Write-Host "🔄 Continuous monitoring mode" -ForegroundColor Yellow
        Write-Host "⏱️  Monitoring interval: $MonitorInterval seconds" -ForegroundColor Yellow
        Write-Host "🛑 Press Ctrl+C to stop" -ForegroundColor Yellow
    }
    
    Write-Host ""
    
    # Set environment variables if needed
    $env:MONITOR_INTERVAL = $MonitorInterval
    $env:MONITOR_DURATION = if ($IsDemoMode) { $MonitorDuration } else { 0 }
    $env:DEMO_MODE = if ($IsDemoMode) { "true" } else { "false" }
    
    try {
        if ($IsDemoMode) {
            # Run with timeout in demo mode
            $job = Start-Job -ScriptBlock {
                param($workDir)
                Set-Location $workDir
                go run cmd/scripts/ssot_performance_monitor.go
            } -ArgumentList (Get-Location)
            
            Write-Host "⏳ Demo mode: Running for $MonitorDuration seconds..." -ForegroundColor Yellow
            
            # Wait for the specified duration or until job completes
            $timeout = New-TimeSpan -Seconds $MonitorDuration
            $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
            
            while ($job.State -eq "Running" -and $stopwatch.Elapsed -lt $timeout) {
                Start-Sleep -Seconds 1
                
                # Show progress every 30 seconds
                if (($stopwatch.Elapsed.TotalSeconds % 30) -eq 0) {
                    $remaining = $MonitorDuration - $stopwatch.Elapsed.TotalSeconds
                    Write-Host "⏰ Time remaining: $([math]::Round($remaining)) seconds" -ForegroundColor Blue
                }
            }
            
            # Stop the job if it's still running
            if ($job.State -eq "Running") {
                Stop-Job $job
                Write-Host "🛑 Demo completed after $MonitorDuration seconds" -ForegroundColor Green
            }
            
            # Get job output
            $output = Receive-Job $job
            if ($output) {
                Write-Host "Job Output:" -ForegroundColor Cyan
                $output | Write-Host
            }
            
            Remove-Job $job
            
        } else {
            # Run normally in continuous mode
            go run cmd/scripts/ssot_performance_monitor.go
        }
        
    } catch {
        Write-Host "❌ Error running performance monitor: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    return $true
}

function Open-Dashboard {
    Write-Host "🌐 Opening Performance Dashboard..." -ForegroundColor Cyan
    
    $dashboardPath = "cmd/scripts/performance_dashboard.html"
    
    if (Test-Path $dashboardPath) {
        $fullPath = Resolve-Path $dashboardPath
        Write-Host "🔗 Dashboard: $fullPath" -ForegroundColor Green
        
        try {
            Start-Process $fullPath
            Write-Host "✅ Dashboard opened in default browser" -ForegroundColor Green
        } catch {
            Write-Host "⚠️  Could not open browser automatically. Please open: $fullPath" -ForegroundColor Yellow
        }
    } else {
        Write-Host "❌ Dashboard file not found: $dashboardPath" -ForegroundColor Red
    }
}

function Show-Summary {
    Write-Host ""
    Write-Host "📈 Performance Monitoring Summary" -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Generated Files:" -ForegroundColor Yellow
    
    # Find generated metric files
    $metricFiles = Get-ChildItem -Name "ssot_metrics_*.json" | Sort-Object -Descending
    
    if ($metricFiles.Count -gt 0) {
        Write-Host "✅ Found $($metricFiles.Count) metric files:" -ForegroundColor Green
        $metricFiles | ForEach-Object {
            $fileSize = (Get-Item $_).Length
            Write-Host "  📄 $_ ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor White
        }
        
        Write-Host ""
        Write-Host "💡 Latest metrics file: $($metricFiles[0])" -ForegroundColor Blue
    } else {
        Write-Host "⚠️  No metric files found" -ForegroundColor Yellow
    }
    
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "• Review generated JSON files for detailed metrics" -ForegroundColor White
    Write-Host "• Open dashboard: .\run_performance_monitor.ps1 -Dashboard" -ForegroundColor White
    Write-Host "• Run continuous monitoring for production use" -ForegroundColor White
    Write-Host ""
}

# Main execution
function Main {
    Clear-Host
    
    Write-Host "🚀 SSOT Performance Monitoring System" -ForegroundColor Magenta
    Write-Host "=====================================" -ForegroundColor Magenta
    Write-Host ""
    
    if ($Help) {
        Show-Help
        return
    }
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        Write-Host "❌ Prerequisites check failed. Please fix issues above." -ForegroundColor Red
        return
    }
    
    # Open dashboard if requested
    if ($Dashboard) {
        Open-Dashboard
        if (-not $Demo) {
            Write-Host ""
            Write-Host "Dashboard opened. Run with -Demo to also start monitoring." -ForegroundColor Blue
            return
        }
    }
    
    # Start monitoring
    $success = Start-PerformanceMonitor -IsDemoMode:$Demo -MonitorDuration:$Duration -MonitorInterval:$Interval
    
    if ($success -and $Demo) {
        Show-Summary
    }
    
    if ($Dashboard -and $Demo) {
        Write-Host "🌐 Dashboard should still be open for viewing results" -ForegroundColor Blue
    }
}

# Run the main function
Main
