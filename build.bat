@echo off
setlocal

echo ========================================
echo HAR Decode Build Script
echo ========================================

:: 设置输出目录
set OUTPUT_DIR=dist

:: 创建输出目录
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

:: 下载依赖
echo.
echo [1/3] Downloading dependencies...
go mod download
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to download dependencies
    exit /b 1
)

:: 整理依赖
echo.
echo [2/3] Tidying dependencies...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to tidy dependencies
    exit /b 1
)

:: 编译
echo.
echo [3/3] Building...
go build -o %OUTPUT_DIR%/har-decode.exe ./cmd/har-decode
if %ERRORLEVEL% neq 0 (
    echo Error: Build failed
    exit /b 1
)

echo.
echo ========================================
echo Build successful!
echo Output: %OUTPUT_DIR%/har-decode.exe
echo ========================================

endlocal
