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

func User_indexID (a, b User) bool {
  return a.ID < b.ID
}

func (p User) IID() int {
  return p.ID
}

func (p User) ToString() string {
  return fmt.Sprintf("%v|%v\n", p.ID, p.Name)
}
