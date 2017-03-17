package main

import "github.com/runner-mei/gengen/types"

func CamelCase(name string) string {
	return types.CamelCase(name)
}

func Underscore(name string) string {
	return types.Underscore(name)
}

func Pluralize(name string) string {
	return types.Pluralize(name)
}

func Singularize(word string) string {
	return types.Singularize(word)
}

func Tableize(className string) string {
	return types.Tableize(className)
}

func Capitalize(word string) string {
	return types.Capitalize(word)
}

func Typeify(word string) string {
	return types.Typeify(word)
}
