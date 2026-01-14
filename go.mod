module github.com/rubiojr/any-vcard

go 1.24.0

replace github.com/rubiojr/anytype-go => ./anytype-go

require (
	github.com/emersion/go-vcard v0.0.0-20230815062825-8fda7d206ec9
	github.com/rubiojr/anytype-go v0.5.0
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.6.1
	golang.org/x/text v0.33.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
