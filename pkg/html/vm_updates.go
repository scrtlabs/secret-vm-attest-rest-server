// pkg/html/vm_updates.go
package html

import (
	"embed"
	"text/template"
)

//go:embed vm_updates.html
var vmUpdatesFS embed.FS

// VMUpdatesTmpl is a parsed template for the VM updates page.
var VMUpdatesTmpl = template.Must(
	template.New("vm_updates.html").ParseFS(vmUpdatesFS, "vm_updates.html"),
)
