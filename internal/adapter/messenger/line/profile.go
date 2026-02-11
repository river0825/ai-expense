package line

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/riverlin/aiexpense/internal/usecase"
)

// LineProfile represents a LINE user profile
type LineProfile struct {
	DisplayName   string `json:"displayName"`
	Language      string `json:"language"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// GetProfile fetches the user's LINE profile including language.
func (c *Client) GetProfile(ctx context.Context, userID string) (*LineProfile, error) {
	url := fmt.Sprintf("%s/%s", c.profileURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.channelToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		lineLogger.ErrorContext(ctx, "failed to fetch LINE profile", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		lineLogger.ErrorContext(
			ctx,
			"LINE profile API returned error",
			"status", resp.StatusCode,
			"body_preview", previewText(string(body), 400),
			"user_id", userID,
		)
		return nil, fmt.Errorf("line profile api error: status %d", resp.StatusCode)
	}

	var profile LineProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	lineLogger.DebugContext(ctx, "LINE profile fetched", "user_id", userID, "language", profile.Language)

	return &profile, nil
}

// LineProfileFetcher adapts the LINE Client to the usecase.ProfileFetcher interface.
type LineProfileFetcher struct {
	client *Client
}

var _ usecase.ProfileFetcher = (*LineProfileFetcher)(nil)

// NewLineProfileFetcher creates a new LineProfileFetcher.
func NewLineProfileFetcher(client *Client) *LineProfileFetcher {
	return &LineProfileFetcher{client: client}
}

// GetLanguage fetches the user's language from their LINE profile.
func (f *LineProfileFetcher) GetLanguage(ctx context.Context, userID string) (string, error) {
	profile, err := f.client.GetProfile(ctx, userID)
	if err != nil {
		return "", err
	}
	return profile.Language, nil
}
