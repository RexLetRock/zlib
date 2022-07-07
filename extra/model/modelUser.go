package model

import (
  "fmt"
)

type User struct {
  ID int
  Name string
  // Extra string
  // DID string
}

func User_Index_Name (a, b User) bool {
  if a.Name < b.Name {
    return true
  }

  if a.Name > b.Name {
    return false
  }

  return a.ID < b.ID
}

func (p User) IID() int {
  return p.ID
}

func (p User) ToString() string {
  return fmt.Sprintf("%v|%v\n", p.ID, p.Name)
}
