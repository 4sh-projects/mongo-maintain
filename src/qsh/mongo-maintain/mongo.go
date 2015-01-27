package main

import "log"
import "bytes"
import "os/exec"
import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"
import "path"
import "time"

var mongoMaintainCollection = "mongomaintain"
var dumpPath = "/tmp/mongomaintain/"
var mongoDefaultArgs []string
var mgoSession *mgo.Session
var database string

type ScriptObject struct {
  ID bson.ObjectId `bson:"_id,omitempty"`
  Script string
  Hash string
  Status string
  Detail string
  Date time.Time
}

func makeScriptObject (name string, hash string) ScriptObject {
  return ScriptObject{"", name, hash, "OK", "", time.Now()}
}

func initMongoContext(params Params) {
  // Build base arguments to pass to mongo
  mongoDefaultArgs = []string{}

  if params.user != "" {
    mongoDefaultArgs = append(mongoDefaultArgs, "-u", params.user)
  }

  if params.password != "" {
    mongoDefaultArgs = append(mongoDefaultArgs, "-p", params.password)
  }

  mongoDefaultArgs = append(mongoDefaultArgs, params.url + "/" + params.database)
  database = params.database
}

func getMongoSession () (*mgo.Session, error) {
  if mgoSession == nil {
    var err error
    mgoSession, err = mgo.Dial(params.url)

    if err != nil {
      return nil , err
    }

    // Manage credentials to connect to mongo
    if params.user != "" {
      credential := mgo.Credential{}
      credential.Username = params.user
      credential.Password = params.password

      mgoSession.Login(&credential)
    }
  }

  return mgoSession.Clone(), nil
}

func queryMongo(query func(*mgo.Collection) error) error {
  session, err := getMongoSession()

  if err != nil {
    return err
  }

  defer session.Close()
  c := session.DB(database).C(mongoMaintainCollection)

  err = query(c)

  return err
}

func getCurrentMongoDumpPath() string {
  return path.Join(dumpPath, database)
}

func mongoDump () error {
  mongoDumpArgs := []string{}

  if params.user != "" {
    mongoDumpArgs = append(mongoDumpArgs, "--username", params.user)
  }

  if params.password != "" {
    mongoDumpArgs = append(mongoDumpArgs, "--password", params.password)
  }

  if params.url != "" {
    mongoDumpArgs = append(mongoDumpArgs, "--host", params.url)
  }

  if params.database != "" {
    mongoDumpArgs = append(mongoDumpArgs, "--db", params.database)
  }

  mongoDumpArgs = append(mongoDumpArgs, "-o", getCurrentMongoDumpPath())

  cmd := exec.Command("mongodump", mongoDumpArgs...)
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()

  if err != nil {
    log.Println(err)
    return err
  }

  log.Printf("Database %s was dumped here %s", database, getCurrentMongoDumpPath())

  return nil
}

func launchMongoScript (filePath string) error {
  args := make([]string, len(mongoDefaultArgs) + 1)
  args = append(mongoDefaultArgs, filePath)

  cmd := exec.Command("mongo", args...)
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()

  return err
}

func tryToGetScriptObjectFromDb (fileName string) (ScriptObject, error) {
  var result ScriptObject

  query := func(collection *mgo.Collection) error {
    return collection.Find(bson.M{"script": fileName}).One(&result)
  }

  err := queryMongo(query)

  return result, err
}

func saveOrUpdateScript (scriptObject ScriptObject) error {
  query := func(collection *mgo.Collection) error {
    scriptObject.Date = time.Now()

    if scriptObject.ID == "" {
      return collection.Insert(&scriptObject)
    } else {
      _, err := collection.UpsertId(scriptObject.ID, &scriptObject)
      return err
    }
  }

  err := queryMongo(query)

  return err
}
