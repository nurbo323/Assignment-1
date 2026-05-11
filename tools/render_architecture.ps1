Add-Type -AssemblyName System.Drawing

$OutputPath = Join-Path $PSScriptRoot '..\architecture\architecture.png'
$OutputPath = [System.IO.Path]::GetFullPath($OutputPath)

$width = 1800
$height = 980
$bitmap = New-Object System.Drawing.Bitmap $width, $height
$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
$graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$graphics.Clear([System.Drawing.ColorTranslator]::FromHtml('#f8fafc'))

$titleFont = New-Object System.Drawing.Font('Arial', 28, [System.Drawing.FontStyle]::Bold)
$subtitleFont = New-Object System.Drawing.Font('Arial', 16, [System.Drawing.FontStyle]::Regular)
$bodyFont = New-Object System.Drawing.Font('Arial', 13, [System.Drawing.FontStyle]::Regular)
$smallFont = New-Object System.Drawing.Font('Arial', 11, [System.Drawing.FontStyle]::Regular)

function New-FillBrush {
    param([string]$Hex)
    return New-Object System.Drawing.SolidBrush([System.Drawing.ColorTranslator]::FromHtml($Hex))
}

function Draw-CenteredText {
    param(
        [string]$Text,
        [System.Drawing.Font]$Font,
        [string]$Color,
        [System.Drawing.RectangleF]$Rect,
        [int]$LineSpacing = 4
    )

    $format = New-Object System.Drawing.StringFormat
    $format.Alignment = [System.Drawing.StringAlignment]::Center
    $format.LineAlignment = [System.Drawing.StringAlignment]::Center
    $brush = New-FillBrush $Color
    $graphics.DrawString($Text, $Font, $brush, $Rect, $format)
    $brush.Dispose()
    $format.Dispose()
}

function Draw-MultiLineBox {
    param(
        [int]$X,
        [int]$Y,
        [int]$W,
        [int]$H,
        [string]$Fill,
        [string]$Border,
        [string]$Title,
        [string[]]$Lines,
        [string]$TitleColor
    )

    $rect = New-Object System.Drawing.Rectangle($X, $Y, $W, $H)
    $fillBrush = New-FillBrush $Fill
    $borderPen = New-Object System.Drawing.Pen([System.Drawing.ColorTranslator]::FromHtml($Border), 3)
    $graphics.FillRectangle($fillBrush, $rect)
    $graphics.DrawRectangle($borderPen, $rect)

    Draw-CenteredText -Text $Title -Font $subtitleFont -Color $TitleColor -Rect ([System.Drawing.RectangleF]::new($X, $Y + 8, $W, 36))

    $lineY = $Y + 56
    foreach ($line in $Lines) {
        Draw-CenteredText -Text $line -Font $bodyFont -Color '#334155' -Rect ([System.Drawing.RectangleF]::new($X + 12, $lineY, $W - 24, 22))
        $lineY += 24
    }

    $fillBrush.Dispose()
    $borderPen.Dispose()
}

function Draw-Arrow {
    param(
        [int]$X1,
        [int]$Y1,
        [int]$X2,
        [int]$Y2,
        [string]$Color
    )

    $pen = New-Object System.Drawing.Pen([System.Drawing.ColorTranslator]::FromHtml($Color), 4)
    $graphics.DrawLine($pen, $X1, $Y1, $X2, $Y2)

    $angle = [Math]::Atan2($Y2 - $Y1, $X2 - $X1)
    $arrowSize = 14
    $p1 = New-Object System.Drawing.Point([int]$X2, [int]$Y2)
    $p2 = New-Object System.Drawing.Point([int]($X2 - $arrowSize * [Math]::Cos($angle - 0.4)), [int]($Y2 - $arrowSize * [Math]::Sin($angle - 0.4)))
    $p3 = New-Object System.Drawing.Point([int]($X2 - $arrowSize * [Math]::Cos($angle + 0.4)), [int]($Y2 - $arrowSize * [Math]::Sin($angle + 0.4)))
    $graphics.FillPolygon((New-FillBrush $Color), @($p1, $p2, $p3))
    $pen.Dispose()
}

function Draw-Label {
    param(
        [int]$X,
        [int]$Y,
        [int]$W,
        [string]$Text
    )

    $rect = New-Object System.Drawing.Rectangle($X, $Y, $W, 24)
    $graphics.FillRectangle((New-FillBrush '#ffffff'), $rect)
    $graphics.DrawRectangle((New-Object System.Drawing.Pen([System.Drawing.ColorTranslator]::FromHtml('#cbd5e1'), 1)), $rect)
    Draw-CenteredText -Text $Text -Font $smallFont -Color '#475569' -Rect ([System.Drawing.RectangleF]::new($X, $Y + 1, $W, 22))
}

$graphics.DrawString('AP2 Assignment 1 / 4 — Microservices Architecture', $titleFont, (New-FillBrush '#0f172a'), 70, 18)
$graphics.DrawString('Clean Architecture • Redis cache • RabbitMQ events • Notification worker', $bodyFont, (New-FillBrush '#334155'), 70, 64)

Draw-MultiLineBox -X 70 -Y 180 -W 250 -H 150 -Fill '#e0f2fe' -Border '#38bdf8' -Title 'Client' -Lines @('Postman / curl', 'creates and checks orders') -TitleColor '#0369a1'
Draw-MultiLineBox -X 380 -Y 145 -W 380 -H 220 -Fill '#eff6ff' -Border '#60a5fa' -Title 'Order Service' -Lines @('HTTP Handler', 'Order Use Case', 'PostgreSQL Repository', 'Redis Cache') -TitleColor '#1d4ed8'
Draw-MultiLineBox -X 840 -Y 145 -W 360 -H 220 -Fill '#f0fdf4' -Border '#34d399' -Title 'Payment Service' -Lines @('HTTP Handler', 'Payment Use Case', 'PostgreSQL Repository', 'RabbitMQ Publisher') -TitleColor '#047857'
Draw-MultiLineBox -X 1260 -Y 170 -W 260 -H 170 -Fill '#fffbeb' -Border '#f59e0b' -Title 'RabbitMQ' -Lines @('payment.completed queue', 'durable', '1 consumer') -TitleColor '#b45309'
Draw-MultiLineBox -X 870 -Y 490 -W 360 -H 220 -Fill '#f5f3ff' -Border '#8b5cf6' -Title 'Notification Service' -Lines @('RabbitMQ Consumer', 'Background Worker', 'Redis state store', 'SIMULATED / SMTP provider') -TitleColor '#6d28d9'
Draw-MultiLineBox -X 1270 -Y 505 -W 250 -H 170 -Fill '#ecfeff' -Border '#14b8a6' -Title 'Redis' -Lines @('lock keys', 'status keys', 'TTL-based cleanup') -TitleColor '#0f766e'
Draw-MultiLineBox -X 420 -Y 500 -W 320 -H 180 -Fill '#fff7ed' -Border '#fb923c' -Title 'Databases' -Lines @('Order DB', 'Payment DB', 'own persistence per service') -TitleColor '#c2410c'

Draw-Arrow -X1 320 -Y1 255 -X2 380 -Y2 255 -Color '#4b5563'
Draw-Arrow -X1 760 -Y1 255 -X2 840 -Y2 255 -Color '#4b5563'
Draw-Arrow -X1 1200 -Y1 255 -X2 1260 -Y2 255 -Color '#4b5563'
Draw-Arrow -X1 1390 -Y1 340 -X2 1050 -Y2 490 -Color '#4b5563'
Draw-Arrow -X1 1030 -Y1 365 -X2 580 -Y2 500 -Color '#4b5563'
Draw-Arrow -X1 1050 -Y1 610 -X2 1270 -Y2 605 -Color '#4b5563'
Draw-Arrow -X1 580 -Y1 365 -X2 580 -Y2 500 -Color '#4b5563'

Draw-Label -X 305 -Y 220 -W 120 -Text 'POST /orders'
Draw-Label -X 710 -Y 220 -W 110 -Text 'payment call'
Draw-Label -X 1210 -Y 220 -W 120 -Text 'publish event'
Draw-Label -X 1210 -Y 400 -W 90 -Text 'consume'
Draw-Label -X 1000 -Y 570 -W 150 -Text 'idempotency / lock'
Draw-Label -X 430 -Y 400 -W 110 -Text 'cache refresh'

$graphics.DrawString('Order Service caches reads in Redis • Payment Service emits events • Notification worker is idempotent via Redis', $smallFont, (New-FillBrush '#475569'), 240, 905)

$bitmap.Save($OutputPath, [System.Drawing.Imaging.ImageFormat]::Png)
$graphics.Dispose()
$bitmap.Dispose()
Write-Host "architecture.png updated at $OutputPath"