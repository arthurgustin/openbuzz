package mocks

import (
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/stretchr/testify/mock"
)

type Client struct {
	mock.Mock
}

func (c *Client) List() []orm.Prospect {
	args := c.Called()
	return args.Get(0).([]orm.Prospect)
}
