package main

import "fmt"

// Data are the values contained by a Feed.
type Data struct {
	ID           int     `json:"id,omitempty"`
	Value        string  `json:"value,omitempty"`
	Position     string  `json:"position,omitempty"`
	FeedID       int     `json:"feed_id,omitempty"`
	GroupID      int     `json:"group_id,omitempty"`
	Expiration   string  `json:"expiration,omitempty"`
	Latitude     float64 `json:"lat,omitempty"`
	Longitude    float64 `json:"lon,omitempty"`
	Elevation    float64 `json:"ele,omitempty"`
	CompletedAt  string  `json:"completed_at,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
	CreatedEpoch float64 `json:"created_epoch,omitempty"`
}

type DataFilter struct {
	StartTime string `url:"start_time,omitempty"`
	EndTime   string `url:"end_time,omitempty"`
}

type DataService struct {
	client *Client
}

// Send adds a new Data value to an existing Feed, or will create the Feed if
// it doesn't already exist.
func (s *DataService) Send(dp *Data) (*Data, *Response, error) {
	path, ferr := s.client.Feed.Path("/data/send")
	if ferr != nil {
		return nil, nil, ferr
	}

	req, rerr := s.client.NewRequest("POST", path, dp)
	if rerr != nil {
		return nil, nil, rerr
	}

	point := &Data{}
	resp, _ := s.client.Do(req, point)
	/*if err != nil {
		return nil, resp, err
	}*/

	return point, resp, nil
}

// Search has the same response format as All, but it accepts optional params
// with which your data can be queried.
func (s *DataService) Search(filter *DataFilter) ([]*Data, *Response, error) {
	path, ferr := s.client.Feed.Path("/data")
	if ferr != nil {
		return nil, nil, ferr
	}

	req, rerr := s.client.NewRequest("GET", path, nil)
	if rerr != nil {
		return nil, nil, rerr
	}

	// request populates Feed slice
	datas := make([]*Data, 0)
	resp, err := s.client.Do(req, &datas)
	if err != nil {
		return nil, resp, err
	}

	return datas, resp, nil
}

// private method for handling the Next, Prev, and Last commands
func (s *DataService) retrieve(command string) (*Data, *Response, error) {
	path, ferr := s.client.Feed.Path(fmt.Sprintf("/data/%v", command))
	if ferr != nil {
		return nil, nil, ferr
	}

	req, rerr := s.client.NewRequest("GET", path, nil)
	if rerr != nil {
		return nil, nil, rerr
	}

	var data Data
	resp, _ := s.client.Do(req, &data)
	/*if err != nil {
		return nil, resp, err
	}*/

	return &data, resp, nil
}

// Last returns the last Data in the stream.
func (s *DataService) Last() (*Data, *Response, error) {
	return s.retrieve("last")
}
