package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"emperror.dev/errors"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/reddit/models"
	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var _ bot.RemoveGuildHandler = (*Plugin)(nil)

func (p *Plugin) RemoveGuild(g int64) error {
	_, err := models.RedditFeeds(models.RedditFeedWhere.GuildID.EQ(g)).UpdateAllG(context.Background(), models.M{
		"disabled": true,
	})
	if err != nil {
		return errors.WrapIf(err, "failed removing reddit feeds")
	}

	return nil
}

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p, &commands.YAGCommand{
		CmdCategory:          commands.CategoryDebug,
		HideFromCommandsPage: true,
		Name:                 "testreddit",
		Description:          "Tests the reddit feeds in this server by checking the specified post",
		HideFromHelp:         true,
		RequiredArgs:         1,
		Arguments: []*dcmd.ArgDef{
			{Name: "post-id", Type: dcmd.String},
		},
		RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {
			pID := data.Args[0].Str()
			if !strings.HasPrefix(pID, "t3_") {
				pID = "t3_" + pID
			}

			resp, err := p.redditClient.LinksInfo([]string{pID})
			if err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				return "Unknown post", nil
			}

			handlerSlow := &PostHandlerImpl{
				Slow:        true,
				ratelimiter: NewRatelimiter(),
			}

			handlerFast := &PostHandlerImpl{
				Slow:        false,
				ratelimiter: NewRatelimiter(),
			}

			err1 := handlerSlow.handlePost(resp[0], data.GuildData.GS.ID)
			err2 := handlerFast.handlePost(resp[0], data.GuildData.GS.ID)

			return fmt.Sprintf("SlowErr: `%v`, fastErr: `%v`", err1, err2), nil
		}),
	})
}

func (p *Plugin) Status() (string, string) {
	numFeeds, err := models.RedditFeeds(models.RedditFeedWhere.Disabled.EQ(false)).CountG(context.Background())
	if err != nil {
		logger.WithError(err).Error("failed fetching status")
		return "Reddit feeds", "error"
	}

	return "Reddit feeds", fmt.Sprintf("%d", numFeeds)
}

// func (p *Plugin) Status() (string, string) {
// 	subs := 0
// 	channels := 0
// 	cursor := "0"

// 	common.

// 	for {
// 		reply := client.Cmd("SCAN", cursor, "MATCH", "global_subreddit_watch:*")
// 		if reply.Err != nil {
// 			logrus.WithError(reply.Err).Error("Error scanning")
// 			break
// 		}

// 		elems, err := reply.Array()
// 		if err != nil {
// 			logrus.WithError(err).Error("Error reading reply")
// 			break
// 		}

// 		if len(elems) < 2 {
// 			logrus.Error("Invalid scan")
// 			break
// 		}

// 		newCursor, err := elems[0].Str()
// 		if err != nil {
// 			logrus.WithError(err).Error("Failed retrieving new cursor")
// 			break
// 		}
// 		cursor = newCursor

// 		list, err := elems[1].List()
// 		if err != nil {
// 			logrus.WithError(err).Error("Failed retrieving list")
// 			break
// 		}

// 		for _, key := range list {
// 			config, err := GetConfig(key)
// 			if err != nil {
// 				logrus.WithError(err).Error("Failed reading global config")
// 				continue
// 			}
// 			if len(config) < 1 {
// 				continue
// 			}
// 			subs++
// 			channels += len(config)
// 		}

// 		if cursor == "" || cursor == "0" {
// 			break
// 		}
// 	}

// 	return "Subs/Channels", fmt.Sprintf("%d/%d", subs, channels)
// }

func SearchSubreddits(data string) (bool, error) {
	var redditStruct struct {
		Data struct {
			Dist     int `json:"dist"`
			Children []struct {
			}
		} `json:"data"`
	}

	query := "https://old.reddit.com/subreddits/search.api?q=" + strings.ToLower(data)
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", UserAgent())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	queryErr := json.Unmarshal([]byte(body), &redditStruct)
	if queryErr != nil {
		return false, err
	}

	if redditStruct.Data.Dist == 0 || len(redditStruct.Data.Children) == 0 {
		return false, nil
	}
	return true, nil
}
