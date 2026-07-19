package schematics

// registerBuiltinRules installs every built-in validator. Names are stable and
// documented in the README.
func registerBuiltinRules(s *Schematics) {
	r := map[string]Rule{
		// strings
		"isString":        ruleIsString,
		"notEmpty":        ruleNotEmpty,
		"email":           ruleEmail,
		"maxLength":       ruleMaxLength,
		"minLength":       ruleMinLength,
		"lengthBetween":   ruleLengthBetween,
		"noSpecialChars":  ruleNoSpecialChars,
		"hasSpecialChars": ruleHasSpecialChars,
		"hasUpper":        ruleHasUpper,
		"hasLower":        ruleHasLower,
		"hasDigit":        ruleHasDigit,
		"isURL":           ruleIsURL,
		"notURL":          ruleNotURL,
		"urlHasHost":      ruleURLHasHost,
		"urlHasQuery":     ruleURLHasQuery,
		"isHTTPS":         ruleIsHTTPS,
		"isUUID":          ruleIsUUID,
		"equals":          ruleEquals,
		"inOptions":       ruleInOptions,
		"matchRegex":      ruleMatchRegex,
		"like":            ruleLike,
		// numbers
		"isNumber":    ruleIsNumber,
		"isInteger":   ruleIsInteger,
		"isFloat":     ruleIsFloat,
		"max":         ruleMax,
		"min":         ruleMin,
		"between":     ruleBetween,
		"positive":    rulePositive,
		"negative":    ruleNegative,
		"nonNegative": ruleNonNegative,
		// dates
		"isDate":      ruleIsDate,
		"beforeNow":   ruleBeforeNow,
		"afterNow":    ruleAfterNow,
		"before":      ruleBefore,
		"after":       ruleAfter,
		"betweenTime": ruleBetweenTime,
		// arrays
		"isArray":        ruleIsArray,
		"maxItems":       ruleMaxItems,
		"minItems":       ruleMinItems,
		"unique":         ruleUnique,
		"itemsInOptions": ruleItemsInOptions,
	}
	for name, fn := range r {
		s.rules[name] = fn
	}
}

// registerBuiltinOperators installs every built-in operator.
func registerBuiltinOperators(s *Schematics) {
	o := map[string]Operator{
		"trim":          opTrim,
		"capitalize":    opCapitalize,
		"upper":         opUpper,
		"lower":         opLower,
		"toString":      opToString,
		"add":           opAdd,
		"subtract":      opSubtract,
		"multiply":      opMultiply,
		"divide":        opDivide,
		"round":         opRound,
		"default":       opDefault,
		"arrayToObject": opArrayToObject,
	}
	for name, fn := range o {
		s.operators[name] = fn
	}
}

// registerBuiltinConditions installs every built-in condition.
func registerBuiltinConditions(s *Schematics) {
	c := map[string]Condition{
		"fieldPresent": condFieldPresent,
		"fieldAbsent":  condFieldAbsent,
		"fieldEquals":  condFieldEquals,
	}
	for name, fn := range c {
		s.conditions[name] = fn
	}
}
