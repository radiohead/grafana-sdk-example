package foo

import (
	"github.com/grafana/thema"
	"time"
	"struct"
	"strings"
)

foo: thema.#Lineage & {
	joinSchema: {
		metadata: {
			{
				uid:               string
				creationTimestamp: time.Time & {
					string
				}
				deletionTimestamp?: time.Time & {
					string
				}
				finalizers: [...string]
				resourceVersion: string
				generation:      >=-9223372036854775808 & <=9223372036854775807 & int
				labels: {
					[string]: string
				}
			}
			{
				[!~"^(uid|creationTimestamp|deletionTimestamp|finalizers|resourceVersion|generation|labels|updateTimestamp|createdBy|updatedBy|extraFields)$"]: string
			}
			updateTimestamp: time.Time & {
				string
			}
			createdBy: string
			updatedBy: string
			extraFields: {
				[string]: _
			}
		}
		spec:            _
		_specIsNonEmpty: spec & struct.MinFields(0)
		status: {
			{
				[string]: _
			}
			#OperatorState: {
				lastEvaluation:    string
				state:             "success" | "in_progress" | "failed"
				descriptiveState?: string
				details?: {
					[string]: _
				}
			}
			operatorStates?: {
				[string]: #OperatorState
			}
			additionalFields?: {
				[string]: _
			}
		}
	}
} & {
	schemas: [{
		version: [0, 0]
		schema: {
			spec: {
				name: string
			}
		}
	}]
	name: strings.ToLower(strings.Replace("Foo", "-", "_", -1))
}