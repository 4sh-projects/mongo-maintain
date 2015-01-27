package main

import "flag"
import "log"
import "os"

type Params struct {
  scripts string
  url string
  database string
  user string
  password string
}

func buildParams() Params {
  params := Params{}

  // Get value from command line
  flag.StringVar(&(params.scripts), "scripts", "", "The absolute path to the folder containing mongo scripts to execute. Do not use ~")
  flag.StringVar(&(params.url), "url", "", "The url to access mongo. Example: localhost:27017")
  flag.StringVar(&(params.database), "database", "", "The database to access mongo. Example: myDatabase")
  flag.StringVar(&(params.user), "user", "", "<Optional> The user to use to access mongo")
  flag.StringVar(&(params.password), "password", "", "<Optional> The password to use to access mongo")

  flag.Parse()

  // Display command line values
  log.Println("* Welcome to the mongo maintain program !!!")
  log.Printf("* -scripts=%s\n", params.scripts)
  log.Printf("* -url=%s\n", params.url)
  log.Printf("* -user=%s\n", params.user)
  log.Printf("* -password=%s\n", params.password)
  log.Println("---------------------------------------------")
  log.Println("")

  // Make some checks
  if params.scripts == "" {
    log.Println("-scripts argument is mandatory")
    os.Exit(-1);
  }

  if params.url == "" {
    log.Println("-url argument is mandatory")
    os.Exit(-1);
  }

  if params.database == "" {
    log.Println("-database argument is mandatory")
    os.Exit(-1);
  }

  return params
}
