# mongo-maintain
Run your scripts in your mongodb with security for each environnement

For users :


For devs :

Install mgo
$go get gopkg.in/mgo.v2


To launch the code
$ go run main.go mongo.go params.go script-file.go utils.go -scripts=./MongoMaintain -url=localhost:27017 -database=test

or

$ export GOPATH=<path to the main folder mongo-maintain>
$ go build
$ ./mongo-maintain -scripts=/Users/dro/Documents/MongoMaintain -url=localhost:27017 -database=test2
