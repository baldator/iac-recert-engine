module test

go 1.24

require (
	github.com/baldator/iac-recert-csvlookup-plugin v0.0.0
	github.com/baldator/iac-recert-engine v0.0.0
	go.uber.org/zap v1.27.0
)

replace github.com/baldator/iac-recert-csvlookup-plugin => ./plugins/csvlookup

replace github.com/baldator/iac-recert-engine => ./iac-recert-engine
