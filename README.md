go-mock
=======

A mock framework for Go (golang).

Usage
-----

Let's say that we have an interface like this that we want to Mock.

	type Client interface {
		Request(url *url.URL) (int, string, error)
	}


We need to create a new struct that implements the interface. But we will use
github.com/maraino/go-mock to replace the actual calls with some specific results.

	import (
		"github.com/maraino/go-mock"
		"net/url"
	)

	type MyClient struct {
		mock.Mock
	}

	func (c *MyClient) Request(url *url.URL) (int, string, error) {
		ret := c.Called(url)
		return ret.Int(0), ret.String(1), ret.Error(2)
	}

Then we need to configure the responses for the defined functions:

	c := &MyClient{}
	url, _ := url.Parse("http://www.example.org")
	c.When("Request", url).Return(200, "{result:1}", nil).Times(1)

We will execute the function that we have Mocked:

	code, json, err := c.Request(url)
	fmt.Printf("Code: %d, JSON: %s, Error: %v\n", code, json, err)

This will produce the output:

	Code: 200, JSON: {result:1}, Error: <nil>

And finally if we want to verify the number of calls we can use:

	if ok, err := c.Verify(); !ok {
		fmt.Println(err)
	}
