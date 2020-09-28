module github.com/iotdomain/ipcam

go 1.13

require (
	github.com/iotdomain/iotdomain-go v0.0.0-20200809060156-51b5ee50e2db
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20200810151505-1b9f1253b3ed // indirect
)

// Temporary for testing iotdomain-go
replace github.com/iotdomain/iotdomain-go => ../iotdomain-go
