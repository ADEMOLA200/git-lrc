package appcore

import (
	"net/http"
	"testing"

	"github.com/HexmosTech/git-lrc/internal/reviewmodel"
)

func TestIsLiveReviewAPIKeyInvalid(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "valid unauthorized code",
			err: &reviewmodel.APIError{
				StatusCode: http.StatusUnauthorized,
				Body:       `{"error_code":"LIVE_REVIEW_API_KEY_INVALID","error":"invalid"}`,
			},
			want: true,
		},
		{
			name: "unauthorized but different error code",
			err: &reviewmodel.APIError{
				StatusCode: http.StatusUnauthorized,
				Body:       `{"error_code":"SOMETHING_ELSE","error":"nope"}`,
			},
			want: false,
		},
		{
			name: "non-401 should not trigger recovery",
			err: &reviewmodel.APIError{
				StatusCode: http.StatusTooManyRequests,
				Body:       `{"error_code":"LIVE_REVIEW_API_KEY_INVALID"}`,
			},
			want: false,
		},
		{
			name: "malformed json should not trigger recovery",
			err: &reviewmodel.APIError{
				StatusCode: http.StatusUnauthorized,
				Body:       `not-json`,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLiveReviewAPIKeyInvalid(tt.err)
			if got != tt.want {
				t.Fatalf("isLiveReviewAPIKeyInvalid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAPIErrorCode(t *testing.T) {
	code, err := parseAPIErrorCode(`{"error_code":"LIVE_REVIEW_API_KEY_INVALID"}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != "LIVE_REVIEW_API_KEY_INVALID" {
		t.Fatalf("unexpected code: %s", code)
	}

	if _, err := parseAPIErrorCode(`{not-json}`); err == nil {
		t.Fatal("expected parse error for malformed json")
	}
}
