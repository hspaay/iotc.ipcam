module github.com/iotdomain/ipcam

go 1.13

require (
	github.com/iotdomain/iotdomain-go v0.0.0-20200809060156-51b5ee50e2db
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
)

// Temporary for testing iotdomain-go
replace github.com/iotdomain/iotdomain-go => ../iotdomain-go
