package main

import "fmt"
import "errors"
import "log"
import "os"
import "crypto/md5"
import "io"
import "encoding/hex"

func computeMd5 (filePath string) (string, error) {
  file, err := os.Open(filePath)
  if err != nil {
    log.Print(err)
    return "", errors.New(fmt.Sprintf("ERROR: error when opening file %s to compute md5", filePath))
  }
  defer file.Close()

  hash := md5.New()
  _, err = io.Copy(hash, file);
  if err != nil {
    log.Print(err)
    return "", errors.New(fmt.Sprintf("ERROR: error when reading file %s", filePath))
  }

  return hex.EncodeToString(hash.Sum(nil)), nil
}
