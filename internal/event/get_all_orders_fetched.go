package event

import "time"

type GetAllOrdersFetched struct {
	Name    string
	Payload interface{}
}

func NewGetAllOrdersFetched() *GetAllOrdersFetched {
	return &GetAllOrdersFetched{
		Name: "GetAllOrdersFetched",
	}
}

func (e *GetAllOrdersFetched) GetName() string {
	return e.Name
}

func (e *GetAllOrdersFetched) GetPayload() interface{} {
	return e.Payload
}

func (e *GetAllOrdersFetched) SetPayload(payload interface{}) {
	e.Payload = payload
}

func (e *GetAllOrdersFetched) GetDateTime() time.Time {
	return time.Now()
}
