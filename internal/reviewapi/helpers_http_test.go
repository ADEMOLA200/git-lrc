package reviewapi

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNewReviewHTTPClient_BlocksCrossHostRedirect(t *testing.T) {
	client := newReviewHTTPClient(5 * time.Second)
	if client.CheckRedirect == nil {
		t.Fatal("expected CheckRedirect to be configured")
	}

	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "attacker.example"}}
	via := []*http.Request{{URL: &url.URL{Scheme: "https", Host: "livereview.hexmos.com"}}}

	err := client.CheckRedirect(req, via)
	if err != http.ErrUseLastResponse {
		t.Fatalf("expected http.ErrUseLastResponse, got %v", err)
	}
}

func TestNewReviewHTTPClient_AllowsSameHostRedirect(t *testing.T) {
	client := newReviewHTTPClient(5 * time.Second)
	if client.CheckRedirect == nil {
		t.Fatal("expected CheckRedirect to be configured")
	}

	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "livereview.hexmos.com"}}
	via := []*http.Request{{URL: &url.URL{Scheme: "https", Host: "livereview.hexmos.com"}}}

	err := client.CheckRedirect(req, via)
	if err != nil {
		t.Fatalf("expected nil for same host redirect, got %v", err)
	}
}
