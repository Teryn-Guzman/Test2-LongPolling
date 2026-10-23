package main

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseObservationRequest(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantNil   bool
		wantAfter int64
		wantWait  time.Duration
		wantErr   bool
	}{
		{name: "prompt snapshot", url: "/v1/jobs/job", wantNil: true},
		{name: "after uses default hold", url: "/v1/jobs/job?after=7", wantAfter: 7, wantWait: 20 * time.Second},
		{name: "explicit hold", url: "/v1/jobs/job?after=0&wait=1", wantAfter: 0, wantWait: time.Second},
		{name: "wait without after", url: "/v1/jobs/job?wait=2", wantErr: true},
		{name: "negative after", url: "/v1/jobs/job?after=-1", wantErr: true},
		{name: "malformed after", url: "/v1/jobs/job?after=abc", wantErr: true},
		{name: "wait too short", url: "/v1/jobs/job?after=0&wait=0", wantErr: true},
		{name: "wait too long", url: "/v1/jobs/job?after=0&wait=21", wantErr: true},
		{name: "unknown query", url: "/v1/jobs/job?mode=long", wantErr: true},
		{name: "duplicate after", url: "/v1/jobs/job?after=1&after=2", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", test.url, nil)
			got, err := parseObservationRequest(request)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseObservationRequest() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if (got == nil) != test.wantNil {
				t.Fatalf("parseObservationRequest() = %#v, wantNil %v", got, test.wantNil)
			}
			if got != nil && (got.after != test.wantAfter || got.wait != test.wantWait) {
				t.Fatalf("parseObservationRequest() = %#v, want after=%d wait=%s", got, test.wantAfter, test.wantWait)
			}
		})
	}
}

func TestJobNotifierBroadcastsToEveryObserver(t *testing.T) {
	notifier := newJobNotifier()
	first, unsubscribeFirst := notifier.subscribe("public-job")
	defer unsubscribeFirst()
	second, unsubscribeSecond := notifier.subscribe("public-job")
	defer unsubscribeSecond()

	notifier.publish("public-job")
	for name, channel := range map[string]<-chan struct{}{"first": first, "second": second} {
		select {
		case <-channel:
		default:
			t.Errorf("%s observer did not receive the broadcast", name)
		}
	}
}

func TestJobNotifierCoalescesNotifications(t *testing.T) {
	notifier := newJobNotifier()
	wake, unsubscribe := notifier.subscribe("public-job")
	defer unsubscribe()

	notifier.publish("public-job")
	notifier.publish("public-job")
	select {
	case <-wake:
	default:
		t.Fatal("observer did not receive notification")
	}
	select {
	case <-wake:
		t.Fatal("notifications should coalesce while observer is rechecking")
	default:
	}
}
