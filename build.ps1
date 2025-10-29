# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

param (
    [string]$Major = "1",
    [string]$Minor = "1",
    [string]$Patch = "0",
    [string]$Out = "out"
)

$ComputeTypes = @("managed-ec2", "anywhere", "managed-containers")
$AppName = "amazon-gamelift-servers-game-server-wrapper"
$AppPackage = "github.com/amazon-gamelift/$AppName"
$GoOS = "windows"
$GoArch = "amd64"
$FileExt = ".exe"
$OutFolder = "$Out\$GoOS\$GoArch\"
$BuildDir = "src"
$ServerSdkUrl = "https://github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/releases/download/v5.4.0/GameLift-Go-ServerSDK-5.4.0.zip"
$ServerSdkFileName = "gamelift-server-sdk.zip"
$ServerSdkExtractDir = "src/ext/gamelift-servers-server-sdk"

function Clean {
    Write-Host "Cleaning output directory..."
    Remove-Item -Recurse -Force $Out -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force $ServerSdkExtractDir -ErrorAction SilentlyContinue
    Remove-Item -Force $ServerSdkFileName -ErrorAction SilentlyContinue
}

function DownloadGameLiftSDK {
    Write-Host "Downloading server SDK for Amazon GameLift Servers..."
    New-Item -ItemType Directory -Path $ServerSdkExtractDir -Force | Out-Null

    if (Test-Path $ServerSdkFileName) {
        Remove-Item -Force $ServerSdkFileName
    }
    if (Test-Path $ServerSdkExtractDir) {
        Remove-Item -Recurse -Force $ServerSdkExtractDir
    }

    Invoke-WebRequest -Uri $ServerSdkUrl -OutFile $ServerSdkFileName
    Expand-Archive -Path $ServerSdkFileName -DestinationPath $ServerSdkExtractDir -Force
}

function Build {
    Clean
    DownloadGameLiftSDK
    Write-Host "Building $AppName for $GoOS/$GoArch..."

    New-Item -ItemType Directory -Path $OutFolder -Force | Out-Null

    $Env:CGO_ENABLED = "0"
    $Env:GOOS = $GoOS
    $Env:GOARCH = $GoArch

    $BuildCmd = @(
        "go build",
        "-C $BuildDir",
        "-trimpath",
        "-v",
        "-ldflags `"",
        "-X '$AppPackage/internal.version=$Major.$Minor.$Patch'",
        "`"",
        "-o ..\$OutFolder\$AppName\$AppName$FileExt",
        "."
    ) -join " "

    Invoke-Expression $BuildCmd

    foreach ($Type in $ComputeTypes) {
        $ComputeTypeOutFolder = "$OutFolder\gamelift-servers-$Type\"

        New-Item -ItemType Directory -Path $ComputeTypeOutFolder -Force | Out-Null

        Copy-Item "$OutFolder\$AppName\$AppName$FileExt" "$ComputeTypeOutFolder\$AppName$FileExt" -Force
        Copy-Item "src\template\template-$Type-config.yaml" "$ComputeTypeOutFolder\config.yaml" -Force
    }

    if ($Type -eq "managed-containers") {
        Copy-Item "src\template\Dockerfile" "$ComputeTypeOutFolder\Dockerfile" -Force
    }

    # Cleanup unnecessary Wrapper folder and ServerSDK file
    Remove-Item -Recurse -Force "$OutFolder\$AppName" -ErrorAction SilentlyContinue
    Remove-Item -Force $ServerSdkFileName
}

function Test {
    Build
    Write-Host "Running tests..."
    Invoke-Expression "go test -C $BuildDir ./..."
}

# Default action
Test
