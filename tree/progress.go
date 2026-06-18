package tree

import "time"

type ProgressEvent struct {
	Where   string
	Iter    int
	Done    int
	Total   int
	Picked  int
	Pending int
	BTS     int
	Note    string
	Time    time.Time
}

type ProgressHook func(ProgressEvent)

func emit(h ProgressHook, e ProgressEvent) {
	if h == nil {
		return
	}
	e.Time = time.Now()
	h(e)
}
