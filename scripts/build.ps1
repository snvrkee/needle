$os_arr = "linux", "darwin", "windows"
$arch_arr = "arm64", "amd64"
$base = "bin"

foreach ($os in $os_arr) {
    $ext = ""
    if ($os -eq "windows") {
        $ext = ".exe"
    }
    foreach ($arch in $arch_arr) {
        $env:GOOS = $os
        $env:GOARCH = $arch
        $name = "{0}/{1}/{2}/needle{3}" -f $base, $os, $arch, $ext
        go build -o $name .
    }
}
