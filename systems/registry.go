package systems

import components "github.com/code-cell/esive/components"

var registry *components.Registry

func SetRegistry(r *components.Registry) {
	registry = r
}
