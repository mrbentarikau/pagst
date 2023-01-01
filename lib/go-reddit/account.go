package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Account contains user account information.
type Account struct {
	IsEmployee           bool   `json:"is_employee"`
	HasVisitedNewProfile bool   `json:"has_visited_new_profile"`
	IsFriend             bool   `json:"is_friend"`
	PrefNoProfanity      bool   `json:"pref_no_profanity"`
	HasExternalAccount   bool   `json:"has_external_account"`
	PrefGeopopular       string `json:"pref_geopopular"`
	PrefShowTrending     bool   `json:"pref_show_trending"`
	Subreddit            struct {
		DefaultSet                 bool          `json:"default_set"`
		UserIsContributor          bool          `json:"user_is_contributor"`
		BannerImg                  string        `json:"banner_img"`
		AllowedMediaInComments     []interface{} `json:"allowed_media_in_comments"`
		UserIsBanned               bool          `json:"user_is_banned"`
		FreeFormReports            bool          `json:"free_form_reports"`
		CommunityIcon              interface{}   `json:"community_icon"`
		ShowMedia                  bool          `json:"show_media"`
		IconColor                  string        `json:"icon_color"`
		UserIsMuted                interface{}   `json:"user_is_muted"`
		DisplayName                string        `json:"display_name"`
		HeaderImg                  interface{}   `json:"header_img"`
		Title                      string        `json:"title"`
		Coins                      int           `json:"coins"`
		PreviousNames              []interface{} `json:"previous_names"`
		Over18                     bool          `json:"over_18"`
		IconSize                   []int         `json:"icon_size"`
		PrimaryColor               string        `json:"primary_color"`
		IconImg                    string        `json:"icon_img"`
		Description                string        `json:"description"`
		SubmitLinkLabel            string        `json:"submit_link_label"`
		HeaderSize                 interface{}   `json:"header_size"`
		RestrictPosting            bool          `json:"restrict_posting"`
		RestrictCommenting         bool          `json:"restrict_commenting"`
		Subscribers                int           `json:"subscribers"`
		SubmitTextLabel            string        `json:"submit_text_label"`
		IsDefaultIcon              bool          `json:"is_default_icon"`
		LinkFlairPosition          string        `json:"link_flair_position"`
		DisplayNamePrefixed        string        `json:"display_name_prefixed"`
		KeyColor                   string        `json:"key_color"`
		Name                       string        `json:"name"`
		IsDefaultBanner            bool          `json:"is_default_banner"`
		URL                        string        `json:"url"`
		Quarantine                 bool          `json:"quarantine"`
		BannerSize                 []int         `json:"banner_size"`
		UserIsModerator            bool          `json:"user_is_moderator"`
		AcceptFollowers            bool          `json:"accept_followers"`
		PublicDescription          string        `json:"public_description"`
		LinkFlairEnabled           bool          `json:"link_flair_enabled"`
		DisableContributorRequests bool          `json:"disable_contributor_requests"`
		SubredditType              string        `json:"subreddit_type"`
		UserIsSubscriber           bool          `json:"user_is_subscriber"`
	} `json:"subreddit"`
	PrefShowPresence    bool        `json:"pref_show_presence"`
	SnoovatarImg        string      `json:"snoovatar_img"`
	SnoovatarSize       interface{} `json:"snoovatar_size"`
	GoldExpiration      interface{} `json:"gold_expiration"`
	HasGoldSubscription bool        `json:"has_gold_subscription"`
	IsSponsor           bool        `json:"is_sponsor"`
	NumFriends          int         `json:"num_friends"`
	Features            struct {
		ModServiceMuteWrites      bool `json:"mod_service_mute_writes"`
		PromotedTrendBlanks       bool `json:"promoted_trend_blanks"`
		ShowAmpLink               bool `json:"show_amp_link"`
		Chat                      bool `json:"chat"`
		IsEmailPermissionRequired bool `json:"is_email_permission_required"`
		ModAwards                 bool `json:"mod_awards"`
		MwebXpromoRevampV3        struct {
			Owner        string `json:"owner"`
			Variant      string `json:"variant"`
			ExperimentID int    `json:"experiment_id"`
		} `json:"mweb_xpromo_revamp_v3"`
		MwebXpromoRevampV2 struct {
			Owner        string `json:"owner"`
			Variant      string `json:"variant"`
			ExperimentID int    `json:"experiment_id"`
		} `json:"mweb_xpromo_revamp_v2"`
		AwardsOnStreams                                bool `json:"awards_on_streams"`
		WebhookConfig                                  bool `json:"webhook_config"`
		MwebXpromoModalListingClickDailyDismissibleIos bool `json:"mweb_xpromo_modal_listing_click_daily_dismissible_ios"`
		LiveOrangereds                                 bool `json:"live_orangereds"`
		CookieConsentBanner                            bool `json:"cookie_consent_banner"`
		ModlogCopyrightRemoval                         bool `json:"modlog_copyright_removal"`
		DoNotTrack                                     bool `json:"do_not_track"`
		ModServiceMuteReads                            bool `json:"mod_service_mute_reads"`
		ChatUserSettings                               bool `json:"chat_user_settings"`
		UsePrefAccountDeployment                       bool `json:"use_pref_account_deployment"`
		MwebXpromoInterstitialCommentsIos              bool `json:"mweb_xpromo_interstitial_comments_ios"`
		ChatSubreddit                                  bool `json:"chat_subreddit"`
		MwebSharingClipboard                           struct {
			Owner        string `json:"owner"`
			Variant      string `json:"variant"`
			ExperimentID int    `json:"experiment_id"`
		} `json:"mweb_sharing_clipboard"`
		PremiumSubscriptionsTable                          bool `json:"premium_subscriptions_table"`
		MwebXpromoInterstitialCommentsAndroid              bool `json:"mweb_xpromo_interstitial_comments_android"`
		CrowdControlForPost                                bool `json:"crowd_control_for_post"`
		NoreferrerToNoopener                               bool `json:"noreferrer_to_noopener"`
		ChatGroupRollout                                   bool `json:"chat_group_rollout"`
		ResizedStylesImages                                bool `json:"resized_styles_images"`
		SpezModal                                          bool `json:"spez_modal"`
		MwebXpromoModalListingClickDailyDismissibleAndroid bool `json:"mweb_xpromo_modal_listing_click_daily_dismissible_android"`
		ExpensiveCoinsPackage                              bool `json:"expensive_coins_package"`
		ActivityServiceRead                                bool `json:"activity_service_read"`
		ActivityServiceWrite                               bool `json:"activity_service_write"`
		AdblockTest                                        bool `json:"adblock_test"`
		AdsAuction                                         bool `json:"ads_auction"`
		AdsAutoExtend                                      bool `json:"ads_auto_extend"`
		AdsAutoRefund                                      bool `json:"ads_auto_refund"`
		AdserverReporting                                  bool `json:"adserver_reporting"`
		AdzerkDoNotTrack                                   bool `json:"adzerk_do_not_track"`
		AdzerkReporting2                                   bool `json:"adzerk_reporting_2"`
		EuCookiePolicy                                     bool `json:"eu_cookie_policy"`
		ExpandoEvents                                      bool `json:"expando_events"`
		ForceHTTPS                                         bool `json:"force_https"`
		GiveHstsGrants                                     bool `json:"give_hsts_grants"`
		HTTPSRedirect                                      bool `json:"https_redirect"`
		ImageUploads                                       bool `json:"image_uploads"`
		ImgurGifConversion                                 bool `json:"imgur_gif_conversion"`
		LegacySearchPref                                   bool `json:"legacy_search_pref"`
		LiveHappeningNow                                   bool `json:"live_happening_now"`
		MoatTracking                                       bool `json:"moat_tracking"`
		MobileNativeBanner                                 bool `json:"mobile_native_banner"`
		MobileSettings                                     bool `json:"mobile_settings"`
		MobileWebTargeting                                 bool `json:"mobile_web_targeting"`
		NewLoggedinCachePolicy                             bool `json:"new_loggedin_cache_policy"`
		NewReportDialog                                    bool `json:"new_report_dialog"`
		OrangeredsAsEmails                                 bool `json:"orangereds_as_emails"`
		OutboundClicktracking                              bool `json:"outbound_clicktracking"`
		PauseAds                                           bool `json:"pause_ads"`
		PostEmbed                                          bool `json:"post_embed"`
		ScreenviewEvents                                   bool `json:"screenview_events"`
		ScrollEvents                                       bool `json:"scroll_events"`
		ShowNewIcons                                       bool `json:"show_new_icons"`
		StickyComments                                     bool `json:"sticky_comments"`
		SubredditRules                                     bool `json:"subreddit_rules"`
		Timeouts                                           bool `json:"timeouts"`
		UpgradeCookies                                     bool `json:"upgrade_cookies"`
		YoutubeScraper                                     bool `json:"youtube_scraper"`
	} `json:"features"`
	CanEditName             bool        `json:"can_edit_name"`
	IsBlocked               bool        `json:"is_blocked"`
	Verified                bool        `json:"verified"`
	NewModmailExists        bool        `json:"new_modmail_exists"`
	PrefAutoplay            bool        `json:"pref_autoplay"`
	Coins                   int         `json:"coins"`
	HasPaypalSubscription   bool        `json:"has_paypal_subscription"`
	HasSubscribedToPremium  bool        `json:"has_subscribed_to_premium"`
	ID                      string      `json:"id"`
	CanCreateSubreddit      bool        `json:"can_create_subreddit"`
	Over18                  bool        `json:"over_18"`
	IsGold                  bool        `json:"is_gold"`
	IsMod                   bool        `json:"is_mod"`
	AwarderKarma            int         `json:"awarder_karma"`
	SuspensionExpirationUtc interface{} `json:"suspension_expiration_utc"`
	HasStripeSubscription   bool        `json:"has_stripe_subscription"`
	IsSuspended             bool        `json:"is_suspended"`
	PrefVideoAutoplay       bool        `json:"pref_video_autoplay"`
	InChat                  bool        `json:"in_chat"`
	HasAndroidSubscription  bool        `json:"has_android_subscription"`
	InRedesignBeta          bool        `json:"in_redesign_beta"`
	IconImg                 string      `json:"icon_img"`
	HasModMail              bool        `json:"has_mod_mail"`
	PrefNightmode           bool        `json:"pref_nightmode"`
	AwardeeKarma            int         `json:"awardee_karma"`
	HideFromRobots          bool        `json:"hide_from_robots"`
	PasswordSet             bool        `json:"password_set"`
	Modhash                 string      `json:"modhash"`
	LinkKarma               int         `json:"link_karma"`
	ForcePasswordReset      bool        `json:"force_password_reset"`
	TotalKarma              int         `json:"total_karma"`
	InboxCount              int         `json:"inbox_count"`
	PrefTopKarmaSubreddits  bool        `json:"pref_top_karma_subreddits"`
	HasMail                 bool        `json:"has_mail"`
	PrefShowSnoovatar       bool        `json:"pref_show_snoovatar"`
	Name                    string      `json:"name"`
	PrefClickgadget         int         `json:"pref_clickgadget"`
	Created                 float64     `json:"created"`
	HasVerifiedEmail        bool        `json:"has_verified_email"`
	GoldCreddits            int         `json:"gold_creddits"`
	CreatedUtc              float64     `json:"created_utc"`
	HasIosSubscription      bool        `json:"has_ios_subscription"`
	PrefShowTwitter         bool        `json:"pref_show_twitter"`
	InBeta                  bool        `json:"in_beta"`
	CommentKarma            int         `json:"comment_karma"`
	AcceptFollowers         bool        `json:"accept_followers"`
	HasSubscribed           bool        `json:"has_subscribed"`
}

// GetMe retrieves the user account for the currently authenticated user. Requires the 'identity' OAuth scope.
func (c *Client) GetMe() (*Account, error) {
	url := fmt.Sprintf("%s/api/v1/me", baseAuthURL)
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

	var account Account
	err = json.NewDecoder(resp.Body).Decode(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
