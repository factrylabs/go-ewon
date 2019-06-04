package dmweb

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(r *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func TestNew(t *testing.T) {
	h := &http.Client{}
	tables := []struct {
		aid string
		u   string
		p   string
		did string
		err error
	}{
		{"", "", "", "", errorMissingCredentials},
		{"accountid", "username", "password", "", errorMissingCredentials},
		{"accountid", "username", "password", "did", nil},
	}

	for _, table := range tables {
		c, err := New(h, table.aid, table.u, table.p, table.did)
		if err != nil {
			assert.Equal(t, table.err, err)
		} else {
			assert.Equal(t, c.baseURL, DefaultBaseURL)
			assert.Equal(t, c.userAgent, DefaultUserAgent)
		}
	}
}

func TestRequest(t *testing.T) {

	tables := []struct {
		aid      string
		u        string
		p        string
		did      string
		endpoint string
		err      error
	}{
		{"accountid", "username", "password", "devid", "getewons", nil},
	}
	for _, table := range tables {
		fc := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		})
		c := &Client{
			Client:    fc,
			AccountID: table.aid,
			Username:  table.u,
			Password:  table.p,
			DevID:     table.did,
			baseURL:   DefaultBaseURL,
			userAgent: DefaultUserAgent}
		res, err := c.Request(table.endpoint, nil)
		assert.NoError(t, err)
		assert.IsType(t, &http.Response{}, res)
		bb, _ := ioutil.ReadAll(res.Body)
		assert.JSONEq(t, string(bb), "{}")
	}
}

func TestGetStatus(t *testing.T) {
	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"historyCount": 20732,
				"ewonsCount": 2,
				"ewons": [{
					"id": 2,
					"name": "Paris",
					"historyCount": 2702,
					"firstHistoryDate": "2015-07-16T16:04:25Z",
					"lastHistoryDate": "2015-07-17T17:43:36Z"
				}, {
					"id": 190,
					"name": "Brussels",
					"historyCount": 18030,
					"firstHistoryDate": "2015-08-24T08:56:44Z",
					"lastHistoryDate": "2015-08-24T14:57:13Z"
				}]
			}`)),
			Header: h,
		}
	})

	// Test successful response
	s, err := c.GetStatus()
	assert.Nil(t, err)
	assert.IsType(t, &GetStatusResponse{}, s)
	assert.Equal(t, 20732, s.HistoryCount)
	assert.Equal(t, 2, s.EwonsCount)
	assert.Equal(t, 2, s.Ewons[0].ID)
	assert.Equal(t, "Paris", s.Ewons[0].Name)
	assert.Equal(t, 2702, s.Ewons[0].HistoryCount)
	fhd, _ := time.Parse(time.RFC3339, "2015-07-16T16:04:25Z")
	assert.Equal(t, fhd, s.Ewons[0].FirstHistoryDate)
	fhd, _ = time.Parse(time.RFC3339, "2015-07-17T17:43:36Z")
	assert.Equal(t, fhd, s.Ewons[0].LastHistoryDate)

	// Test 401 response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 401,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"success":false,"code":401,"message":"Invalid credentials"}`)),
			Header:     h,
		}
	})
	_, err = c.GetStatus()
	assert.Error(t, err)
	assert.Equal(t, "Invalid credentials", err.Error())

}

func TestGetEwons(t *testing.T) {

	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response without timezone
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"success":true,"ewons":[{"id":123456,"name":"Ewon1","lastSynchroDate":"2017-07-08T10:51:28Z"},{"id":123457,"name":"Ewon2","lastSynchroDate":"2018-06-05T12:49:27Z"}]}`)),
			Header:     h,
		}
	})
	es, err := c.GetEwons()
	assert.Nil(t, err)
	assert.IsType(t, Ewons{}, es)
	if assert.Len(t, es, 2) {
		assert.Equal(t, 123456, es[0].ID)
		assert.Equal(t, "Ewon1", es[0].Name)
		t1, _ := time.Parse(time.RFC3339, "2017-07-08T10:51:28Z")
		assert.Equal(t, t1, es[0].LastSynchroDate)
	}

	// Test valid response with timezone
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"success":true,"ewons":[{"id":123456,"name":"Ewon1","timeZone":"Europe/Brussels", "lastSynchroDate":"2017-07-08T10:51:28Z"}]}`)),
			Header:     h,
		}
	})
	es, err = c.GetEwons()
	assert.Nil(t, err)
	assert.IsType(t, Ewons{}, es)
	if assert.Len(t, es, 1) {
		assert.Equal(t, "Europe/Brussels", es[0].TimeZone)
	}

	// Test 401 response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 401,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"success":false,"code":401,"message":"Invalid credentials"}`)),
			Header:     h,
		}
	})
	_, err = c.GetEwons()
	assert.Error(t, err)
	assert.Equal(t, "Invalid credentials", err.Error())
}

func TestGetEwonByID(t *testing.T) {

	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"id": 123456,
				"name": "Ewon1",
				"tags": [
						{
								"id": 98765,
								"name": "Random_Metric",
								"dataType": "Float",
								"description": "",
								"alarmHint": "",
								"value": 1234.4567,
								"quality": "good",
								"ewonTagId": 10
						}
					],
					"lastSynchroDate":"2018-06-05T12:49:27Z"
				}`)),
			Header: h,
		}
	})
	e, err := c.GetEwonByID(123456)
	assert.Nil(t, err)
	assert.IsType(t, &Ewon{}, e)
	assert.Equal(t, 123456, e.ID)
	assert.Equal(t, "Ewon1", e.Name)
	t1, _ := time.Parse(time.RFC3339, "2018-06-05T12:49:27Z")
	assert.Equal(t, t1, e.LastSynchroDate)
	assert.Equal(t, "Random_Metric", e.Tags[0].Name)

	// Test unknown EwonID
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 404,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
			"success": false,
			"code": 404,
			"message": "No eWON found for id '987654'"
	}`)),
			Header: h,
		}
	})
	e, err = c.GetEwonByID(987654)
	assert.Error(t, err)
	assert.Equal(t, "No eWON found for id '987654'", err.Error())
}

func TestGetEwonByName(t *testing.T) {

	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"id": 123456,
				"name": "Ewon1",
				"tags": [
						{
								"id": 98765,
								"name": "Random_Metric",
								"dataType": "Float",
								"description": "",
								"alarmHint": "",
								"value": 1234.4567,
								"quality": "good",
								"ewonTagId": 10
						}
					],
					"lastSynchroDate":"2018-06-05T12:49:27Z"
				}`)),
			Header: h,
		}
	})
	e, err := c.GetEwonByName("Ewon1")
	assert.Nil(t, err)
	assert.IsType(t, &Ewon{}, e)
	assert.Equal(t, 123456, e.ID)
	assert.Equal(t, "Ewon1", e.Name)
	t1, _ := time.Parse(time.RFC3339, "2018-06-05T12:49:27Z")
	assert.Equal(t, t1, e.LastSynchroDate)
	assert.Equal(t, "Random_Metric", e.Tags[0].Name)
}

func TestGetData(t *testing.T) {
	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response with ewonID filter
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		assert.Equal(t, "508238", req.URL.Query().Get("ewonId"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"moreDataAvailable": true,
				"ewons": [{
					"id": 508238,
					"name": "ltn_flexy",
					"tags": [{
						"id": 780591,
						"name": "TAG_2",
						"dataType": "Float",
						"description": "",
						"alarmHint": "",
						"value": 1510,
						"quality": "good",
						"ewonTagId": 2,
						"history": [{
								"date": "2018-11-08T14:17:58Z",
								"quality": "initialGood",
								"value": 0
							},
							{
								"date": "2018-11-08T14:18:00Z",
								"value": 0
							},
							{
								"date": "2018-11-08T14:18:02Z",
								"value": 0
							}, {
								"date": "2018-11-08T14:18:04Z",
								"value": 0
							},
							{
								"date": "2018-11-08T14:18:06Z",
								"value": 0
							}
						]
					}],
					"lastSynchroDate": "2018-11-09T09:47:00Z",
					"timeZone": "Europe/Brussels"
				}]
			}`)),
			Header: h,
		}
	})

	params := make(map[string]string)
	params["ewonId"] = "508238"
	d, err := c.GetData(params)
	assert.Nil(t, err)
	assert.IsType(t, &GetDataResponse{}, d)
	assert.Equal(t, true, d.MoreDataAvailable)
}

func TestSyncData(t *testing.T) {
	c := &Client{
		AccountID: "aid",
		Username:  "username",
		Password:  "password",
		DevID:     "devid",
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}

	// Test valid response
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"transactionId": "456789",
				"moreDataAvailable": true,
				"ewons": [{
					"id": 508238,
					"name": "ltn_flexy",
					"tags": [{
						"id": 780591,
						"name": "TAG_2",
						"dataType": "Float",
						"description": "",
						"alarmHint": "",
						"value": 1510,
						"quality": "good",
						"ewonTagId": 2,
						"history": [{
							"date": "2018-11-08T14:17:58Z",
							"quality": "initialGood",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:00Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:02Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:04Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:06Z",
							"value": 0
						}]
					}],
					"lastSynchroDate": "2018-11-09T09:47:00Z",
					"timeZone": "Europe/Brussels"
				}]
			}`)),
			Header: h,
		}
	})
	s, err := c.FirstSyncData()
	assert.Nil(t, err)
	assert.IsType(t, &SyncResponse{}, s)
	assert.Equal(t, true, s.MoreDataAvailable)
	assert.Equal(t, "456789", s.TransactionID)

	// Test valid response with lastTransactionID
	c.Client = NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "go-ewon/dmweb 0.1", req.Header.Get("User-Agent"))
		h := make(http.Header)
		h.Add("Content-Type", "application/json;charset=UTF-8")
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`{
				"success": true,
				"transactionId": "987654",
				"moreDataAvailable": false,
				"ewons": [{
					"id": 508238,
					"name": "ltn_flexy",
					"tags": [{
						"id": 780591,
						"name": "TAG_2",
						"dataType": "Float",
						"description": "",
						"alarmHint": "",
						"value": 1510,
						"quality": "good",
						"ewonTagId": 2,
						"history": [{
							"date": "2018-11-08T14:17:58Z",
							"quality": "initialGood",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:00Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:02Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:04Z",
							"value": 0
						}, {
							"date": "2018-11-08T14:18:06Z",
							"value": 0
						}]
					}],
					"lastSynchroDate": "2018-11-09T09:47:00Z",
					"timeZone": "Europe/Brussels"
				}]
			}`)),
			Header: h,
		}
	})
	s, err = c.SyncData(s.TransactionID, true)
	assert.Nil(t, err)
	assert.IsType(t, &SyncResponse{}, s)
	assert.Equal(t, false, s.MoreDataAvailable)
	assert.Equal(t, "987654", s.TransactionID)

}
