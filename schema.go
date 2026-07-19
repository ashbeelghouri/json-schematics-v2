package schematics

// Schema is the declarative document that drives validation and operation.
//
// Example (JSON):
//
//	{
//	  "version": "2.0",
//	  "separator": ".",
//	  "arrayIdKey": "id",
//	  "locale": "en",
//	  "db": { "minAge": 18 },
//	  "fields": [
//	    {
//	      "target": "user.profile.name",
//	      "type": "string",
//	      "required": true,
//	      "validate": [ { "rule": "minLength", "args": { "min": 2 } } ],
//	      "operate":  [ { "op": "trim" }, { "op": "capitalize" } ]
//	    }
//	  ]
//	}
type Schema struct {
	Version    string         `json:"version,omitempty"`
	Separator  string         `json:"separator,omitempty"`
	ArrayIDKey string         `json:"arrayIdKey,omitempty"`
	Locale     string         `json:"locale,omitempty"`
	DB         map[string]any `json:"db,omitempty"`
	Fields     []Field        `json:"fields"`
}

// Field targets one or more flattened keys and describes how to validate and
// transform the matched values.
type Field struct {
	// Target selects flattened keys: a literal path, a path with a "*"
	// wildcard, or (when TargetRegex is true) a regular expression.
	Target      string `json:"target"`
	TargetRegex bool   `json:"targetRegex,omitempty"`

	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`

	// Required fails when the target selects no value.
	Required bool `json:"required,omitempty"`
	// DependsOn lists other targets that must be present for this field to be
	// validated.
	DependsOn []string `json:"dependsOn,omitempty"`
	// AddToDB copies the matched value into the shared DB before validation, so
	// other rules can reference it.
	AddToDB bool `json:"addToDB,omitempty"`

	Tags []string `json:"tags,omitempty"`

	// When lists conditions that must all hold for the field to be processed.
	When []ConditionRef `json:"when,omitempty"`
	// Validate is the ordered chain of validators. The first failing rule stops
	// evaluation of this field and produces one error.
	Validate []RuleRef `json:"validate,omitempty"`
	// Operate is the ordered chain of operators applied by Operate.
	Operate []OperatorRef `json:"operate,omitempty"`

	Meta map[string]any `json:"meta,omitempty"`
}

// RuleRef references a registered validator and its per-use configuration.
type RuleRef struct {
	Rule     string            `json:"rule"`
	Args     Args              `json:"args,omitempty"`
	Message  string            `json:"message,omitempty"`
	Messages map[string]string `json:"messages,omitempty"`
}

// OperatorRef references a registered operator and its arguments.
type OperatorRef struct {
	Op   string `json:"op"`
	Args Args   `json:"args,omitempty"`
}

// ConditionRef references a registered condition. Negate inverts its result.
type ConditionRef struct {
	Condition string `json:"condition"`
	Args      Args   `json:"args,omitempty"`
	Negate    bool   `json:"negate,omitempty"`
}
