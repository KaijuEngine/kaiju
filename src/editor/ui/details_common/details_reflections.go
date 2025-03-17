package details_common

func IsNumber(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	default:
		return false
	}
}

func IsInput(typeName string) bool {
	return typeName == "string" || IsNumber(typeName)
}

func IsCheckbox(typeName string) bool {
	return typeName == "bool"
}

func IsEntityId(packageName, typeName string) bool {
	return packageName == "kaiju/engine" && typeName == "EntityId"
}
