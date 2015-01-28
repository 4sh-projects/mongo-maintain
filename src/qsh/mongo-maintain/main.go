package main

import "log"
import "os"
import "io/ioutil"
import "regexp"
import "sort"
import "path"

var params Params
var fileNameValidator = regexp.MustCompile(`^v[\d]*([\._][\d]*)*__.*\.js$`)

func main() {
  params = buildParams()

  scripts := []ScriptFile{}
  files, _ := ioutil.ReadDir(params.scripts)

  // Iterate on files contained in 'scripts' folder
  for _, f := range files {
    if !f.IsDir() {
      scriptFile, err := makeScriptFile(f.Name())

      if err != nil {
        log.Println(err)
        stop()
      }

      scripts = append(scripts, scriptFile);
    }
  }

  initMongoContext(params)
  dumpErr := mongoDump()

  if dumpErr != nil {
    log.Println("ERROR Fail to create dump before to treat scripts")
    stop()
  }

  sort.Sort(ByVersion(scripts))

  for _, script := range scripts {
    filePath := path.Join(params.scripts, script.name)

    scriptObject, err := tryToGetScriptObjectFromDb(script.name)

    if (err != nil && err.Error() != "not found") {
      log.Println(err)
      log.Printf("ERROR when trying to get script %s from db\n", script.name)
      stopBecauseOfFailure()
    }

    hashString, err := computeMd5(filePath)

    if err != nil {
      log.Println(err)
      log.Printf("ERROR during md5 computing of script %s\n", script.name)
      stopBecauseOfFailure()
    }

    if err == nil && scriptObject.Status == "OK" {
      if hashString != scriptObject.Hash {
        log.Printf("ERROR: %s was already launched, but hash is not the same\n", script.name)
        stopBecauseOfFailure()
      } else {
        log.Printf("- %s was already launched\n", script.name)
      }
    } else {
      if (err != nil && "not found" == err.Error()) {
        log.Println("not found")
        scriptObject = makeScriptObject(script.name, hashString)
      }

      log.Println(scriptObject)

      err = launchMongoScript(filePath)

      if err != nil {
        log.Printf("- %s launching FAILED\n", script.name)
        log.Println(err)
        manageScriptFailure(script.name, hashString, "ERROR when launching script\n")
      }

      scriptObject.Status = "OK"

      err = saveOrUpdateScript(scriptObject)

      if err != nil {
        log.Printf("- %s was successfully launched but INFORMATION WAS NOT SAVED IN DB\n", script.name)
        log.Println(err)
        stopBecauseOfFailure()
      }

      log.Printf("- %s was successfully launched and saved\n", script.name)
    }
  }
}

func stop () {
  os.Exit(-1);
}

func stopBecauseOfFailure () {
  log.Printf("Because of an error the program was interrupted.")
  log.Printf("A dump of your database %s was created here %s before any modification. You can restore it with mongorestore", database, getCurrentMongoDumpPath())

  stop()
}

func manageScriptFailure (name string, hash string, detail string) {
  scriptObject := makeScriptObject(name, hash)
  scriptObject.Status = "KO"
  scriptObject.Detail = detail

  err := saveOrUpdateScript(scriptObject)

  if err != nil {
    log.Printf("- %s FAILED but impossible to save it db\n", name)
  }

  stopBecauseOfFailure()
}
