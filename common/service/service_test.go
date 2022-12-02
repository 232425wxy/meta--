package service

import (
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBaseService_Start(t *testing.T) {
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	bs := NewBaseService(logger, "Blockchain")

	err := bs.Start()
	assert.Nil(t, err)

	err = bs.Start()
	assert.Equal(t, ErrAlreadyStarted, err)
}

func TestBaseService_WaitStop(t *testing.T) {
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	bs := NewBaseService(logger, "Blockchain")
	err := bs.Start()
	assert.Nil(t, err)

	err = bs.Stop()
	assert.Nil(t, err)

	err = bs.Stop()
	assert.Equal(t, ErrAlreadyStopped, err)
}
