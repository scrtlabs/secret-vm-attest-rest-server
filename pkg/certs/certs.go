package certs

import _ "embed"

//go:embed sectigo-r46.crt
var IntelRootCA []byte
