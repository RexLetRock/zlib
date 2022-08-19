package main

import (
  "bufio"
  "os"
  "fmt"
  "crypto/sha1"

  "github.com/google/uuid"
  "github.com/bwmarrin/snowflake"

  // "github.com/RexLetRock/zlib/zlog"
  "github.com/RexLetRock/zlib/zbench"

  // "github.com/RexLetRock/zlib/extra/model"
)

var (
  NRun = 1_000_000
  NCpu = 12
  node, _ = snowflake.NewNode(int64(10))
)

func main() {
  benchZID()

  fmt.Printf("\n\nStop with ctrl + c \n\n")
  input := bufio.NewScanner(os.Stdin)
  input.Scan()
}

func benchZID() {
  fmt.Printf("\n\n=== ZCOUNT ===\n")
  fmt.Printf("\n== UUID RUN %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    id := uuid.New()
    id.String()
  })

  fmt.Printf("\n== SNOWFLAKE RUN %v threads\n", NCpu)
  zbench.Run(NRun, NCpu, func(i, _ int) {
    GenerateToken()
  })
}


func GenerateToken() string {
	hash := sha1.Sum(node.Generate().Bytes())
	token := fmt.Sprintf("%x", hash[:])
	return token
}
