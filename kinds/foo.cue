package kinds

foo: {
	codegen: {
		frontend: true
		backend:  true
	}

	group:   "grafana-sdk-example"
	kind:    "Foo"
	current: "v1alpha1"

	apiResource: {
		scope: "Namespaced"
	}

	versions: {
		v1alpha1: {
			schema: {
				spec: {
					// Name is the name of a Foo.
					name: string
				}
			}
		}
	}
}
