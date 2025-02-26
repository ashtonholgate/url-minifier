package domain

import (
	"strings"
	"testing"
)

func TestShortener_GenerateShortCode(t *testing.T) {
	s := NewShortener()
	
	tests := []struct {
		name   string
		url    string
		userID string
		want   struct {
			length int
			valid  bool
		}
	}{
		{
			name:   "Basic URL",
			url:    "https://example.com",
			userID: "user1",
			want: struct {
				length int
				valid  bool
			}{
				length: ShortCodeLength,
				valid:  true,
			},
		},
		{
			name:   "Same URL different user",
			url:    "https://example.com",
			userID: "user2",
			want: struct {
				length int
				valid  bool
			}{
				length: ShortCodeLength,
				valid:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.GenerateShortCode(tt.url, tt.userID)

			// Check length
			if len(got) != tt.want.length {
				t.Errorf("GenerateShortCode() length = %v, want %v", len(got), tt.want.length)
			}

			// Check if all characters are valid base62
			for _, c := range got {
				if !tt.want.valid || strings.IndexRune(Base62Chars, c) == -1 {
					t.Errorf("GenerateShortCode() contains invalid character: %c", c)
				}
			}

			// Check if different users get different codes for same URL
			if tt.name == "Same URL different user" {
				otherCode := s.GenerateShortCode(tt.url, "user1")
				if got == otherCode {
					t.Error("GenerateShortCode() should generate different codes for different users")
				}
			}
		})
	}
}

func TestShortener_ValidateCustomAlias(t *testing.T) {
	s := NewShortener()

	tests := []struct {
		name    string
		alias   string
		wantErr bool
	}{
		{
			name:    "Valid alias",
			alias:   "my-custom-url",
			wantErr: false,
		},
		{
			name:    "Too short",
			alias:   "ab",
			wantErr: true,
		},
		{
			name:    "Too long",
			alias:   "this-is-a-very-very-very-long-custom-alias",
			wantErr: true,
		},
		{
			name:    "Invalid characters",
			alias:   "my@custom#url",
			wantErr: true,
		},
		{
			name:    "Valid with numbers",
			alias:   "custom-url-123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateCustomAlias(tt.alias)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCustomAlias() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
