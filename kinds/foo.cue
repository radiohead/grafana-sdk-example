package kinds

// This is our Foo definition, which contains metadata about the kind, and the kind's schema
foo: {
	// Name is the human-readable name which is used for generated type names.
	name: "Foo"
	// Group affects the grouping of this kind, especially with respect to API server expression.
	// This is typically the same as the plugin ID, though there are length restrictions when the crd trait is present,
	// so it may vary if the plugin ID is too long.
	group: "grafana-sdk-example"
	// crd is a trait that governs both whether this kind is expressible as a resource in an API server
	// (the `resource.Object` interface), and attributes of that expression.
	// if crd is not present, the kind system and code generation do not impose API server format restricitons,
	// and do not add common API server information.
	crd: {
		// [OPTIONAL]
		// Scope determines the scope of the kind in the API server. It currently allows two values:
		// * Namespaced - resources for this kind are created inside namespaces
		// * Cluster - resource for this kind are always cluster-wide (this can be thought of as a "global" namespace)
		// If not present, this defaults to "Namespaced"
		scope: "Namespaced"
	}
	// [OPTIONAL]
	// Codegen is a trait that tells the grafana-app-sdk, or other code generation tooling, how to process this kind.
	// If not present, default values within the codegen trait are used.
	codegen: {
		// [OPTIONAL]
		// frontend tells the CLI to generate front-end code (TypeScript interfaces) for the schema.
		// Will default to true if not present.
		frontend: true
		// [OPTIONAL]
		// backend tells the CLI to generate backend-end code (Go types) for the schema.
		// Will default to true if not present.
		backend: true
	}
	// [OPTIONAL]
	// The human-readable plural form of the "name" field.
	// Will default to <name>+"s" if not present.
	pluralName: "Foos"
	// lineage is the thema Lineage (Schema) definition.
	// A Lineage is a Thema construct that is used to present versions of a resource in a continuous sequence,
	// with translations between versions. Read more at [https://github.com/grafana/thema].
	lineage: {
		// The lineage's schemas field is the sequence of all schema versions, starting with [0,0].
		// Each entry in the list is a subsequent version, either a new major version ([+1,0]) or a new minor version ([+0,+1])
		schemas: [{
			// Version is the version of this specific schema within the lineage.
			// The first entry must be [0,0], and subsequent ones must increment either the major or minor version.
			// For breaking changes (such as adding a new non-optional field), the major version must be incremented.
			version: [0, 0]
			// Schema is the actual schema for this version
			// As an API server-expressable resource, the schema has a restricted format:
			// {
			//     spec: { ... }
			//     status: { ... } // optional
			//     metadata: { ... } // optional
			// }
			// `spec` must always be present, and is the schema for the object.
			// `status` is optional, and should contain status or state information which is typically not user-editable
			// (controlled by controllers/operators). The kind system adds some implicit status information which is
			// common across all kinds, and becomes present in the unified lineage used for code generation and other tooling.
			// `metadata` is optional, and should contain kind- or schema-specific metadata. The kind system adds
			// an explicit set of common metadata which can be found at in the kindsys public repository at:
			// https://github.com/grafana/grafana-app-sdk/kindsys/blob/452481b6348225a1bdb02c9abaef25d29ffe680d/kindcat_custom.cue#L25
			// additional metadata fields cannot conflict with the kindsys common metadata
			schema: {
				// spec is the schema of our resource. The spec should include all the user-ediable information for the kind.
				spec: {
					name: string
				}
				// status is where state and status information which may be used or updated by the operator or back-end should be placed
				// If you do not have any such information, you do not need to include this field,
				// however, as mentioned above, certain fields will be added by the kind system regardless.
				//status: {
				//	currentState: string
				//}
				// metadata if where kind- and schema-specific metadata goes. This is typically unused,
				// as the kind system's common metadata is always part of `metadata` and covers most metadata use-cases.
				//metadata: {
				//	kindSpecificField: string
				//}
			}
		}]
	}
}
