package mocks

import (
	"github.com/stretchr/testify/mock"
	"open-buzz/orm"
)

type Client struct{
	mock.Mock
}

func (c *Client) List() []orm.Prospect {
	args := c.Called()
	return args.Get (0).([]orm.Prospect)
}
