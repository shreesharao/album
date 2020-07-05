function LogInfo 
{
	Param ([string] $msg)
	Write-Host '--------------------------------------------------'
	Write-Host $msg #-ForegroundColor Black -BackgroundColor Yellow
	Write-Host '--------------------------------------------------'
}

function LogError 
{
	Param ([string] $msg)
	Write-Host '--------------------------------------------------'
	Write-Host $msg -ForegroundColor Black -BackgroundColor Red
	Write-Host '--------------------------------------------------'
}

function LogSuccess 
{
	#Write-Host '--------------------------------------------------'
	#Param ([string] $msg)
	#Write-Host $msg -ForegroundColor Black -BackgroundColor Green
	#Write-Host '--------------------------------------------------'
} 
LogInfo 'Filess cleanup - starting' 

@(
	'api'
	'api.exe'
) |
Where-Object { Test-Path $_ }|
	ForEach-Object { Remove-Item $_ -Recurse -Force -ErrorAction Stop }     
	

if($? -eq $false)
{
	LogError 'Files cleanup - failed'
	exit
}

LogSuccess 'Files cleanup - completed'

LogInfo "Setting environment variables GOOS=linux and CGO_Enabled=0"
set-item env:GOOS "linux"
set-item env:CGO_ENABLED 0
set-item env:GOARCH amd64

go env

LogInfo 'Publishing artifacts'
go build ..\cmd\api\ 

if($?  -eq $false)
{
	LogError 'go build  failed'
	exit
}

LogSuccess 'Publishing artifacts - completed'
$repo = 'album'

LogInfo 'Docker build'
docker build -t ${repo}:latest -f ./Dockerfile .

if($?  -eq $false)
{
	LogError 'docker build failed'
	exit
}

LogSuccess 'Docker build - completed'