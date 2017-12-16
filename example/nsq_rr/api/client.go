package api

type transport interface {
	Call(req interface{}, rsp interface{}) error
	Close()
}

type Client struct {
	t transport
}

func NewClient(r transport) *Client {
	return &Client{t: r}
}

func (c *Client) Add(x, y int) (int, error) {
	req := addReq{X: x, Y: y}
	rsp := &addRsp{}
	if err := c.t.Call(req, rsp); err != nil {
		return 0, err
	}
	return rsp.Z, nil
}

func (c *Client) Close() {
	c.t.Close()
}
