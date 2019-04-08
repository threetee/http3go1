package common

import (
  "os"
)

func FileExists(dir string) bool {
  info, err := os.Stat(dir)
  if err != nil {
    return false
  }

  return !info.IsDir()
}
