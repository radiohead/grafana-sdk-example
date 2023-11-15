//
// Code generated by grafana-app-sdk. DO NOT EDIT.
//

package v1alpha1

import (
	"github.com/grafana/grafana-app-sdk/resource"
)

// schema is unexported to prevent accidental overwrites
var schema = resource.NewSimpleSchema("grafana-sdk-example.ext.grafana.com", "v1alpha1", &Object{}, resource.WithKind("Foo"),
	resource.WithPlural("foos"), resource.WithScope(resource.NamespacedScope))

// Schema returns a resource.SimpleSchema representation of Foo
func Schema() *resource.SimpleSchema {
	return schema
}