package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetSubredditInfo(where string) (*SubredditInfoData, error) {
	url := fmt.Sprintf("%s/r/%s/about.json", baseURL, where)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result *SubredditInfo
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type SubredditInfo struct {
	Kind string            `json:"kind"`
	Data SubredditInfoData `json:"data"`
}

type SubredditInfoData struct {
	UserFlairBackgroundColor       interface{}   `json:"user_flair_background_color"`
	SubmitTextHTML                 string        `json:"submit_text_html"`
	RestrictPosting                bool          `json:"restrict_posting"`
	UserIsBanned                   bool          `json:"user_is_banned"`
	FreeFormReports                bool          `json:"free_form_reports"`
	WikiEnabled                    bool          `json:"wiki_enabled"`
	UserIsMuted                    bool          `json:"user_is_muted"`
	UserCanFlairInSr               interface{}   `json:"user_can_flair_in_sr"`
	DisplayName                    string        `json:"display_name"`
	HeaderImg                      string        `json:"header_img"`
	Title                          string        `json:"title"`
	AllowGalleries                 bool          `json:"allow_galleries"`
	IconSize                       []int         `json:"icon_size"`
	PrimaryColor                   string        `json:"primary_color"`
	ActiveUserCount                int           `json:"active_user_count"`
	IconImg                        string        `json:"icon_img"`
	DisplayNamePrefixed            string        `json:"display_name_prefixed"`
	AccountsActive                 int           `json:"accounts_active"`
	PublicTraffic                  bool          `json:"public_traffic"`
	Subscribers                    int           `json:"subscribers"`
	UserFlairRichtext              []interface{} `json:"user_flair_richtext"`
	Name                           string        `json:"name"`
	Quarantine                     bool          `json:"quarantine"`
	HideAds                        bool          `json:"hide_ads"`
	PredictionLeaderboardEntryType string        `json:"prediction_leaderboard_entry_type"`
	EmojisEnabled                  bool          `json:"emojis_enabled"`
	AdvertiserCategory             string        `json:"advertiser_category"`
	PublicDescription              string        `json:"public_description"`
	CommentScoreHideMins           int           `json:"comment_score_hide_mins"`
	AllowPredictions               bool          `json:"allow_predictions"`
	UserHasFavorited               bool          `json:"user_has_favorited"`
	UserFlairTemplateID            interface{}   `json:"user_flair_template_id"`
	CommunityIcon                  string        `json:"community_icon"`
	BannerBackgroundImage          string        `json:"banner_background_image"`
	OriginalContentTagEnabled      bool          `json:"original_content_tag_enabled"`
	CommunityReviewed              bool          `json:"community_reviewed"`
	SubmitText                     string        `json:"submit_text"`
	DescriptionHTML                string        `json:"description_html"`
	SpoilersEnabled                bool          `json:"spoilers_enabled"`
	CommentContributionSettings    struct {
		AllowedMediaTypes interface{} `json:"allowed_media_types"`
	} `json:"comment_contribution_settings"`
	AllowTalks                       bool          `json:"allow_talks"`
	HeaderSize                       []int         `json:"header_size"`
	UserFlairPosition                string        `json:"user_flair_position"`
	AllOriginalContent               bool          `json:"all_original_content"`
	HasMenuWidget                    bool          `json:"has_menu_widget"`
	IsEnrolledInNewModmail           interface{}   `json:"is_enrolled_in_new_modmail"`
	KeyColor                         string        `json:"key_color"`
	CanAssignUserFlair               bool          `json:"can_assign_user_flair"`
	Created                          float64       `json:"created"`
	Wls                              int           `json:"wls"`
	ShowMediaPreview                 bool          `json:"show_media_preview"`
	SubmissionType                   string        `json:"submission_type"`
	UserIsSubscriber                 bool          `json:"user_is_subscriber"`
	AllowedMediaInComments           []interface{} `json:"allowed_media_in_comments"`
	AllowVideogifs                   bool          `json:"allow_videogifs"`
	ShouldArchivePosts               bool          `json:"should_archive_posts"`
	UserFlairType                    string        `json:"user_flair_type"`
	AllowPolls                       bool          `json:"allow_polls"`
	CollapseDeletedComments          bool          `json:"collapse_deleted_comments"`
	EmojisCustomSize                 interface{}   `json:"emojis_custom_size"`
	PublicDescriptionHTML            string        `json:"public_description_html"`
	AllowVideos                      bool          `json:"allow_videos"`
	IsCrosspostableSubreddit         interface{}   `json:"is_crosspostable_subreddit"`
	NotificationLevel                interface{}   `json:"notification_level"`
	ShouldShowMediaInCommentsSetting bool          `json:"should_show_media_in_comments_setting"`
	CanAssignLinkFlair               bool          `json:"can_assign_link_flair"`
	AccountsActiveIsFuzzed           bool          `json:"accounts_active_is_fuzzed"`
	AllowPredictionContributors      bool          `json:"allow_prediction_contributors"`
	SubmitTextLabel                  string        `json:"submit_text_label"`
	LinkFlairPosition                string        `json:"link_flair_position"`
	UserSrFlairEnabled               bool          `json:"user_sr_flair_enabled"`
	UserFlairEnabledInSr             bool          `json:"user_flair_enabled_in_sr"`
	AllowChatPostCreation            bool          `json:"allow_chat_post_creation"`
	AllowDiscovery                   bool          `json:"allow_discovery"`
	AcceptFollowers                  bool          `json:"accept_followers"`
	UserSrThemeEnabled               bool          `json:"user_sr_theme_enabled"`
	LinkFlairEnabled                 bool          `json:"link_flair_enabled"`
	DisableContributorRequests       bool          `json:"disable_contributor_requests"`
	SubredditType                    string        `json:"subreddit_type"`
	SuggestedCommentSort             interface{}   `json:"suggested_comment_sort"`
	BannerImg                        string        `json:"banner_img"`
	UserFlairText                    interface{}   `json:"user_flair_text"`
	BannerBackgroundColor            string        `json:"banner_background_color"`
	ShowMedia                        bool          `json:"show_media"`
	ID                               string        `json:"id"`
	UserIsModerator                  bool          `json:"user_is_moderator"`
	Over18                           bool          `json:"over18"`
	HeaderTitle                      string        `json:"header_title"`
	Description                      string        `json:"description"`
	IsChatPostFeatureEnabled         bool          `json:"is_chat_post_feature_enabled"`
	SubmitLinkLabel                  string        `json:"submit_link_label"`
	UserFlairTextColor               interface{}   `json:"user_flair_text_color"`
	RestrictCommenting               bool          `json:"restrict_commenting"`
	UserFlairCSSClass                interface{}   `json:"user_flair_css_class"`
	AllowImages                      bool          `json:"allow_images"`
	Lang                             string        `json:"lang"`
	WhitelistStatus                  string        `json:"whitelist_status"`
	URL                              string        `json:"url"`
	CreatedUtc                       float64       `json:"created_utc"`
	BannerSize                       []int         `json:"banner_size"`
	MobileBannerImage                string        `json:"mobile_banner_image"`
	UserIsContributor                bool          `json:"user_is_contributor"`
	AllowPredictionsTournament       bool          `json:"allow_predictions_tournament"`
}
