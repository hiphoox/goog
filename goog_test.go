package goog

import (
  "testing"
  l4g "code.google.com/p/log4go"
)

const (
  logConfigurationFile = "config/goog.xml"
)

func init() {
  l4g.LoadConfiguration(logConfigurationFile)
}

const (
  server = "192.168.1.107:2480"
  database_name = "geneology"
  login = "root"
  password = "957AB56F4CDCE148CEBDD2A06B569073927416B7A2EF95DFFFF7E28E77FEA22A"
)

func TestConnect(t *testing.T) {
    _, err := Connect(server, database_name, login, password)
    if err != nil {
      t.Fatal("Failed test: %s\n", err.Error())
    }
}