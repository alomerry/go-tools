package test

import (
  "sync"
  
  "github.com/stretchr/testify/suite"
)

type BaseSuite struct {
  suite.Suite
  mutex                        sync.Mutex
}

func (suite *BaseSuite) SetupTest() {
}

func (suite *BaseSuite) TearDownTest() {
}

func (suite *BaseSuite) Setup() {
}

func (suite *BaseSuite) TearDown() {

}