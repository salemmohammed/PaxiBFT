package PaxiBFT

import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/salemmohammed/PaxiBFT/lib"
	"github.com/salemmohammed/PaxiBFT/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
)

// Client interface provides get and put for key value store
type Client interface {
	PutMUL(Key, Value) error
	Put(Key, Value) error
}

// AdminClient interface provides fault injection opeartion
type AdminClient interface {
	Consensus(Key) bool
	Crash(ID, int)
	Drop(ID, ID, int)
	Partition(int, ...ID)
}

// HTTPClient inplements Client interface with REST API
type HTTPClient struct {
	Addrs  map[ID]string
	HTTP   map[ID]string
	ID     ID  // client id use the same id as servers in local site
	N      int // total number of nodes
	LocalN int // number of nodes in local zone
	Limit  int
	CID int // command id
	Count  int
	*http.Client
	MyList []ID
}

// NewHTTPClient creates a new Client from config
func NewHTTPClient(id ID) *HTTPClient {
	c := &HTTPClient{
		ID:     id,
		N:      len(config.Addrs),
		Addrs:  config.Addrs,
		HTTP:   config.HTTPAddrs,
		Client: &http.Client{},
		Limit:  0,
		Count:  0,
	}
	if id != "" {
		i := 0
		for node := range c.Addrs {
			if node.Zone() == id.Zone() {
				i++
			}
		}
		c.LocalN = i
	}

	return c
}
func (c *HTTPClient) Put(key Key, value Value) error {
	c.CID++
	_, _, err := c.RESTPut(c.ID, key, value)
	return err
}
func (c *HTTPClient) PutMUL(key Key, value Value) error {
	log.Debugf("<----------------PutMUL---------------->")
	i  := 0
	errs := make(chan error, 0)
	c.Count++
	for id := range c.HTTP {
		//if i>1{
		//	continue
		//}
		log.Debugf("range id %v", id)
		//c.MyList = append(c.MyList,id)
		go func(id ID) {
			c.CID++
			_, _, err := c.rest(id, key, value,c.CID)
			if err != nil {
				log.Error(err)
				return
			}
			errs <- err
		}(id)
		i++;
	}
	errors := make([]error, 0)
	for ; i > 0; i-- {
	errors = append(errors, <-errs)
	}
	log.Debugf("errors %v ", errors)
	//log.Debugf("c.MyList %v ", c.MyList)
	fmt.Println("----------------Done PutMUL---------------->")
	return errors[0]
}
func (c *HTTPClient) GetURL(id ID, key Key) string {
	if id == "" {
		for id = range c.HTTP {
			if c.ID == "" || id.Zone() == c.ID.Zone() {
				break
			}
		}
	}
	return c.HTTP[id] + "/" + strconv.Itoa(int(key))
}

// rest accesses server's REST API with url = http://ip:port/key
// if value == nil, it's a read
func (c *HTTPClient) rest(id ID, key Key, value Value,count int) (Value, map[string]string, error) {

	url := c.GetURL(id, key)
	log.Debugf("In REST %v", id)
	method := http.MethodGet
	var body io.Reader
	if value != nil {
		method = http.MethodPut
		log.Debugf("value=%v",value)
		body = bytes.NewBuffer(value)
		//log.Debugf("method=%v",method)
	}
	//log.Debugf("URL %v", url)
	//time.Sleep(40*time.Millisecond)
	req, err := http.NewRequest(method, url, body)
	//log.Debugf("req=%v",req)
	if err != nil {
		log.Debugf("we cannot create a request")
		log.Error(err)
		return nil, nil, err
	}
	//log.Debugf("HTTPClientID=%v",HTTPClientID)
	//log.Debugf("HTTPCommandID=%v",HTTPCommandID)
	//log.Debugf("c.ID=%v",c.ID)
	//log.Debugf("CID=%v",count)

	req.Header.Set(HTTPClientID, string(id))
	req.Header.Set(HTTPCommandID, strconv.Itoa(c.CID))
	// r.Header.Set(HTTPTimestamp, strconv.FormatInt(time.Now().UnixNano(), 10))

	rep, err := c.Client.Do(req)
	//log.Debugf("c.Client.Do(req)=%v",rep)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}


	defer rep.Body.Close()

	// get headers
	metadata := make(map[string]string)
	for k := range rep.Header {
		metadata[k] = rep.Header.Get(k)
	}
	//log.Debugf("metadata=%v",metadata)
	//log.Debugf("rep.StatusCode=%v",rep.StatusCode)

	if rep.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			log.Error(err)
			return nil, metadata, err
		}

		if value == nil{
			log.Debugf("node=%v type=%s key=%v value=%x", id, method, key, Value(b))
		} else {
			log.Debugf("node=%v type=%s key=%v value=%x", id, method, key, value)
		}
		return Value(b), metadata, nil
	}

	// http call failed
	dump, _ := httputil.DumpResponse(rep, true)
	log.Debugf("%q", dump)
	return nil, metadata, errors.New(rep.Status)
}
// RESTPut puts new value as http.request body and return previous value
func (c *HTTPClient) RESTPut(id ID, key Key, value Value) (Value, map[string]string, error) {
	return c.rest(id, key, value,c.CID)
}

func (c *HTTPClient) json(id ID, key Key, value Value) (Value, error) {
	url := c.HTTP[id]
	cmd := Command{
		Key:       key,
		Value:     value,
		ClientID:  c.ID,
		CommandID: c.CID,
	}
	data, err := json.Marshal(cmd)
	res, err := c.Client.Post(url, "json", bytes.NewBuffer(data))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		log.Debugf("key=%v value=%x", key, Value(b))
		return Value(b), nil
	}
	dump, _ := httputil.DumpResponse(res, true)
	log.Debugf("%q", dump)
	return nil, errors.New(res.Status)
}

// JSONGet posts get request in json format to server url
func (c *HTTPClient) JSONGet(key Key) (Value, error) {
	return c.json(c.ID, key, nil)
}

// JSONPut posts put request in json format to server url
func (c *HTTPClient) JSONPut(key Key, value Value) (Value, error) {
	return c.json(c.ID, key, value)
}

// QuorumGet concurrently read values from majority nodes
func (c *HTTPClient) QuorumGet(key Key) ([]Value, []map[string]string) {
	return c.MultiGet(c.N/2+1, key)
}

// MultiGet concurrently read values from n nodes
func (c *HTTPClient) MultiGet(n int, key Key) ([]Value, []map[string]string) {
	valueC := make(chan Value)
	metaC := make(chan map[string]string)
	i := 0
	for id := range c.HTTP {
		go func(id ID) {
			v, meta, err := c.rest(id, key, nil,c.CID)
			if err != nil {
				log.Error(err)
				return
			}
			valueC <- v
			metaC <- meta
		}(id)
		i++
		if i >= n {
			break
		}
	}

	values := make([]Value, 0)
	metas := make([]map[string]string, 0)
	for ; i > 0; i-- {
		values = append(values, <-valueC)
		metas = append(metas, <-metaC)
	}
	return values, metas
}

func (c *HTTPClient) LocalQuorumGet(key Key) ([]Value, []map[string]string) {
	valueC := make(chan Value)
	metaC := make(chan map[string]string)
	i := 0
	for id := range c.HTTP {
		if c.ID.Zone() != id.Zone() {
			continue
		}
		i++
		if i > c.LocalN/2 {
			break
		}
		go func(id ID) {
			v, meta, err := c.rest(id, key, nil,c.CID)
			if err != nil {
				log.Error(err)
				return
			}
			valueC <- v
			metaC <- meta
		}(id)
	}

	values := make([]Value, 0)
	metas := make([]map[string]string, 0)
	for ; i >= 0; i-- {
		values = append(values, <-valueC)
		metas = append(metas, <-metaC)
	}
	return values, metas
}

// QuorumPut concurrently write values to majority of nodes
// TODO get headers
func (c *HTTPClient) QuorumPut(key Key, value Value) {
	var wait sync.WaitGroup
	i := 0
	for id := range c.HTTP {
		i++
		if i > c.N/2 {
			break
		}
		wait.Add(1)
		go func(id ID) {
			c.rest(id, key, value,c.CID)
			wait.Done()
		}(id)
	}
	wait.Wait()
}

// Consensus collects /history/key from every node and compare their values
func (c *HTTPClient) Consensus(k Key) bool {
	h := make(map[ID][]Value)
	for id, url := range c.HTTP {
		h[id] = make([]Value, 0)
		r, err := c.Client.Get(url + "/history?key=" + strconv.Itoa(int(k)))
		if err != nil {
			log.Error(err)
			continue
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			continue
		}
		holder := h[id]
		err = json.Unmarshal(b, &holder)
		if err != nil {
			log.Error(err)
			continue
		}
		h[id] = holder
		log.Debugf("node=%v key=%v h=%v", id, k, holder)
	}
	n := 0
	for _, v := range h {
		if len(v) > n {
			n = len(v)
		}
	}
	for i := 0; i < n; i++ {
		set := make(map[string]struct{})
		for id := range c.HTTP {
			if len(h[id]) > i {
				set[string(h[id][i])] = struct{}{}
			}
		}
		if len(set) > 1 {
			return false
		}
	}
	return true
}

// Crash stops the node for t seconds then recover
// node crash forever if t < 0
func (c *HTTPClient) Crash(id ID, t int) {
	url := c.HTTP[id] + "/crash?t=" + strconv.Itoa(t)
	r, err := c.Client.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	r.Body.Close()
}

// Drop drops every message send for t seconds
func (c *HTTPClient) Drop(from, to ID, t int) {
	url := c.HTTP[from] + "/drop?id=" + string(to) + "&t=" + strconv.Itoa(t)
	r, err := c.Client.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	r.Body.Close()
}

// Partition cuts the network between nodes for t seconds
func (c *HTTPClient) Partition(t int, nodes ...ID) {
	s := lib.NewSet()
	for _, id := range nodes {
		s.Add(id)
	}
	for from := range c.Addrs {
		if !s.Has(from) {
			for _, to := range nodes {
				c.Drop(from, to, t)
			}
		}
	}
}