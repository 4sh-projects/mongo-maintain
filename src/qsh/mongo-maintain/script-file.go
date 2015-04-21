package main

import "os"
import "log"
import "fmt"
import "errors"
import "strconv"
import "strings"
import "sort"
import "path/filepath"
import "regexp"

var fileNameValidator = regexp.MustCompile(`^v[\d]*([\._][\d]*)*__.*\.js$`)

type ScriptFile struct {
  name string
  path string
  version []string
}

func getScriptFilesFromFolder (folder string) []ScriptFile {
  scripts := []ScriptFile{}

  err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
      // Exclude folder and hiden files
      if (!f.IsDir() && !strings.HasPrefix(f.Name(), ".")) {
        scriptFile, err := makeScriptFile(f.Name(), path)

        if err != nil {
          log.Println(err)
          stop()
        }

        scripts = append(scripts, scriptFile);
      }

      return nil
  })

  if err != nil {
    log.Println(err)
    stop()
  }

  sort.Sort(ByVersion(scripts))

  return scripts
}


func makeScriptFile(name string, path string) (ScriptFile, error) {
  // Check file name validity.
  if !fileNameValidator.MatchString(name) {
    return ScriptFile{}, errors.New(fmt.Sprintf("ERROR: %s is not a valid script name\n", name))
  }

  // Here filename is valid
  parts := strings.Split(name, "__");
  versionPart := parts[0]
  versionPart = versionPart[1:len(versionPart)] // Remove the first char which is 'v'
  versionPart = strings.Replace(versionPart, ".", "_", -1) // Replace all '.' by '_'

  return ScriptFile{name, path, strings.Split(versionPart, "_")}, nil
}

type ByVersion []ScriptFile

func (a ByVersion) Len() int           { return len(a) }
func (a ByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool {
  size := len(a[i].version)

  if (size > len(a[j].version)) {
    size = len(a[j].version)
  }

  for k := 0; k < size; k++ {
    n1, _ := strconv.ParseInt(a[i].version[k], 10, 64)
    n2, _ := strconv.ParseInt(a[j].version[k], 10, 64)

    if (n1 != n2) {
      return n1 < n2
    }
  }

  return true
}
