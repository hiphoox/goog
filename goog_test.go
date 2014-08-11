package goog

import (
  "testing"
  "fmt"
  l4g "code.google.com/p/log4go"
)

const (
  logConfigurationFile = "config/goog.xml"
)

func init() {
  l4g.LoadConfiguration(logConfigurationFile)
}

const (
  server = "localhost:2480"
  database_name = "geneology"
  login = "root"
  password = "957AB56F4CDCE148CEBDD2A06B569073927416B7A2EF95DFFFF7E28E77FEA22A"
)

func TestConnect(t *testing.T) {
    client, err := Connect(server, database_name, login, password)
    if err != nil {
      t.Error("Failed test:", err.Error(), "type:", fmt.Sprintf("%T", err.Error()), "\n")
    }

    token :=  client.GetToken()
    l4g.Info(token)
}