/*
	By default this test suite assumes there is a local Exasol instance
	listening on port 8563 and with a default sys password. You can
	override this via --host, --port, and --pass test arguments.

	We recommend using an Exasol docker container for this:
		https://github.com/exasol/docker-db

	The routines in this file are shared by all the test files.
	There aren't any actual tests in this file.
*/
package exasol

import (
	"flag"
	"github.com/stretchr/testify/suite"
	"testing"
)

var testHost = flag.String("host", "127.0.0.1", "Exasol hostname")
var testPort = flag.Int("port", 8563, "Exasol port")
var testPass = flag.String("pass", "exasol", "Exasol SYS password")
var testLoglevel = flag.String("loglevel", "warning", "Output loglevel")

type testSuite struct {
	suite.Suite
	exaConn  *Conn
	loglevel string
	schema   string
}

func TestExasolClient(t *testing.T) {
	s := initTestSuite()
	s.connectExasol()
	defer s.exaConn.Disconnect()
	suite.Run(t, s)
}

func initTestSuite() *testSuite {
	s := new(testSuite)
	s.loglevel = *testLoglevel
	s.schema = "[test]"
	return s
}

func (s *testSuite) connectExasol() {
	s.exaConn = Connect(ConnConf{
		Host:     *testHost,
		Port:     uint16(*testPort),
		Username: "SYS",
		Password: *testPass,
		LogLevel: s.loglevel,
		Timeout:  10,
	})
}

func (s *testSuite) SetupTest() {
	if s.exaConn != nil {
		s.execute("DROP SCHEMA IF EXISTS " + s.schema + " CASCADE")
		s.execute("CREATE SCHEMA " + s.schema)
	}
}

func (s *testSuite) TearDownTest() {
	if s.exaConn != nil {
		s.exaConn.Rollback()
	}
}

func (s *testSuite) execute(args ...string) {
	for _, arg := range args {
		_, err := s.exaConn.Execute(arg)
		if !s.exaConn.Conf.SuppressError {
			s.NoError(err, "Unable to execute SQL")
		}
	}
}

func (s *testSuite) fetch(sql string) [][]interface{} {
	data, err := s.exaConn.FetchSlice(sql)
	if !s.exaConn.Conf.SuppressError {
		s.NoError(err, "Unable to execute SQL")
	}
	return data
}
