package dmweb

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// DefaultBaseURL is the default URL to access the EWON service.
const DefaultBaseURL = "https://data.talk2m.com/"

// DefaultUserAgent is this package's default User-Agent for making
// requests to EWONs services.
const DefaultUserAgent = "go-ewon/dmweb 0.1"

// parseTime parses eWon times
// Before firmware 13.2, the eWON is always logging data in local time.
// As of firmware 13.2, the eWON has the option to record data using UTC timestamps.
// affects func parseTime()?

// outgoing calls to servers should accept a Context

var (
	errorMissingCredentials    = errors.New("missing one or more credentials")
	errorCouldNotParseArgument = errors.New("could not parse argument")
)

// New constructs a new DMWeb Client
func New(h *http.Client, accountID, username, password, developerID string) (*Client, error) {
	if accountID == "" || username == "" || password == "" || developerID == "" {
		return nil, errorMissingCredentials
	}
	c := Client{
		Client:    h,
		AccountID: accountID,
		Username:  username,
		Password:  password,
		DevID:     developerID,
		baseURL:   DefaultBaseURL,
		userAgent: DefaultUserAgent,
	}
	return &c, nil
}

// Request perform the actual request
func (c *Client) Request(endpoint string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.buildURL(endpoint, params), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", c.userAgent)
	res, err := c.Client.Do(req)
	if err != nil {
		return res, err
	}
	if res.StatusCode != 200 {
		var er errorResponse
		err := json.NewDecoder(res.Body).Decode(&er)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(er.Message)
	}
	return res, err
}

func (c *Client) buildURL(endpoint string, params url.Values) string {
	v := url.Values{}
	v.Add("t2maccount", c.AccountID)
	v.Add("t2musername", c.Username)
	v.Add("t2mpassword", c.Password)
	v.Add("t2mdevid", c.DevID)
	for p, vals := range params {
		for _, val := range vals {
			v.Add(p, val)
		}
	}
	return c.baseURL + endpoint + "?" + v.Encode()
}

// GetStatus returns the storage consumption of the account and of each eWON.
func (c *Client) GetStatus() (*GetStatusResponse, error) {
	res, err := c.Request("getstatus", nil)
	if err != nil {
		return nil, err
	}
	var s GetStatusResponse
	err = json.NewDecoder(res.Body).Decode(&s)
	return &s, err
}

// GetEwons returns all eWons
// The "getewons" service returns the list of eWONs sending data to be stored in the DataMailbox.
// The result contains the following information for each eWON:
// - its name and id,
// - its number of tags, (according to the docs, not in reality)
// - the date of its last data upload to the Data Mailbox.
func (c *Client) GetEwons() (Ewons, error) {
	res, err := c.Request("getewons", nil)
	if err != nil {
		return nil, err
	}
	var es struct {
		Success bool
		Ewons   Ewons
	}
	err = json.NewDecoder(res.Body).Decode(&es)
	return es.Ewons, err
}

func (c *Client) getEwonByIdentifier(qp string, i interface{}) (*Ewon, error) {
	qs := url.Values{}
	switch i.(type) {
	case int:
		qs.Add(qp, strconv.Itoa(i.(int)))
	case string:
		qs.Add(qp, i.(string))
	default:
		return nil, errorCouldNotParseArgument
	}
	res, err := c.Request("getewon", qs)
	if err != nil {
		return nil, err
	}
	var e Ewon
	err = json.NewDecoder(res.Body).Decode(&e)
	return &e, err
}

// GetEwonByID returns a single eWon by ID
func (c *Client) GetEwonByID(id int) (*Ewon, error) {
	return c.getEwonByIdentifier("id", id)
}

// GetEwonByName returns a single eWon by Name
// Name of the eWON as returned by the “getewons” API request.
func (c *Client) GetEwonByName(name string) (*Ewon, error) {
	return c.getEwonByIdentifier("name", name)
}

// GetData is used as a “one-shot” request to retrieve filtered
// data based on specific criteria. It is not destined to grab
// historical data with the same timestamp or enormous data
// involving the use of the "moreData" filter.
// Valid params are:
//   * ewonId: ID of the single eWON for which data from DataMailbox is requested.
//   * tagId: ID of the single tag for which data from DataMailbox is requested.
//   * from: Timestamp after which data should be returned. No data older than this time stamp will be sent.
//   * to: Timestamp before which data should be returned. No data newer than this time stamp will be sent.
//   * fullConfig: By default, getdata returns configuration information only for eWONs / tags that contain historical data.
//     If the request contains "fullConfig" as parameter, all tags / eWONs will appear in the data set, even if they do not
//     contain historical data. "fullConfig" doesn’t accept any value. It is used as is.
//   * limit: The maximum amount of historical data returned.
// If the size of the historical data saved in the DataMailbox exceeds this limit, only the oldest historical data will be returned and the result contains a moreDataAvailable value indicating that more data is available on the server.If the limit parameter is not used or is too high, the DataMailbox uses a limit pre-defined in the system.
func (c *Client) GetData(params map[string]string) (*GetDataResponse, error) {
	qs := url.Values{}
	for k, v := range params {
		qs.Add(k, v)
	}
	res, err := c.Request("getdata", qs)
	if err != nil {
		return nil, err
	}
	var d GetDataResponse
	err = json.NewDecoder(res.Body).Decode(&d)
	return &d, err
}

// FirstSyncData should be used the first time we're syncing data.
// After that, use the SyncData function.
func (c *Client) FirstSyncData() (*SyncResponse, error) {
	return c.SyncData("", true)
}

// SyncData is used to retrieve all the data. This service is
// destined to grab the whole set of data regardless the amount.
// The "syncdata" service retrieves all data of a Talk2M account
// incrementally. Therefore, only new data is returned on each
// API request.
// You can filter the request with the following parameters:
//   * lastTransactionID: The ID of the last set of data sent by
//     the DataMailbox. By referencing the "lastTransactionId",
//     the DataMailbox will send a set of data more recent than
//     the data linked to this transaction ID.
//   * createTransaction: The indication to the server that a
//     new transaction ID should be created for this request.
func (c *Client) SyncData(lastTransactionID string, createTransaction bool) (*SyncResponse, error) {
	qs := url.Values{}
	if lastTransactionID != "" {
		qs.Add("lastTransactionId", lastTransactionID)
	}
	if createTransaction {
		qs.Add("createTransaction", "true")
	}
	res, err := c.Request("syncdata", qs)
	if err != nil {
		return nil, err
	}
	var s SyncResponse
	err = json.NewDecoder(res.Body).Decode(&s)
	return &s, err
}
