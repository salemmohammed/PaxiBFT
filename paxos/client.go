package paxos
import (
	"github.com/salemmohammed/PaxiBFT"
)
type Client struct {
	*PaxiBFT.HTTPClient
	ballot PaxiBFT.Ballot
}
func NewClient(id PaxiBFT.ID) *Client {
	return &Client{
		HTTPClient: PaxiBFT.NewHTTPClient(id),
	}
}
func (c *Client) Put(key PaxiBFT.Key, value PaxiBFT.Value) error {
	c.HTTPClient.CID++
	_, meta, err := c.RESTPut(c.ID, key, value)
	if err == nil {
		b := PaxiBFT.NewBallotFromString(meta[HTTPHeaderBallot])
		if b > c.ballot {
			c.ballot = b
		}
	}

	return err
}