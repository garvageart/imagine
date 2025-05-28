package utils

import "testing"

const (
	AppName = "imagine"
***REMOVED***

var (
	IsProduction  = IsEnvironment("production"***REMOVED***
	IsTest        = testing.Testing(***REMOVED*** || IsEnvironment("test"***REMOVED***
	IsDevelopment = !IsProduction && !IsTest

	Environment = func(***REMOVED*** string {
		switch true {
		case IsProduction:
			return "production"
		case IsTest:
			return "test"
		default:
			return "development"
	***REMOVED***
***REMOVED***(***REMOVED***
***REMOVED***
